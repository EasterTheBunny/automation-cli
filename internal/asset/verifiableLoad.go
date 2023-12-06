package asset

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/montanaflynn/stats"

	verifiableLogTrigger "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_log_trigger_upkeep_wrapper"
	verifiableConditional "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_upkeep_wrapper"
)

var (
	ErrContractRead       = fmt.Errorf("contract read error")
	ErrCalculationFailure = fmt.Errorf("calculation failure")
	//nolint:gochecknoglobals
	headerRowOne = table.Row{
		"ID", "Total Performs", "Block Delays", "Block Delays", "Block Delays",
		"Block Delays", "Block Delays", "Block Delays", "Block Delays"}
	//nolint:gochecknoglobals
	headerRowTwo = table.Row{
		"", "", "50th", "90th", "95th", "99th", "Max", "Total", "Average"}
)

const (
	P50                   float64 = 50
	P90                   float64 = 90
	P95                   float64 = 95
	P99                   float64 = 99
	percentilesCalculated         = 4
	// workerNum is the total number of workers calculating upkeeps' delay summary
	workerNum = 20
	// retryDelay is the time the go routine will wait before calling the same contract function
	retryDelay = 1 * time.Second
	// retryNum defines how many times the go routine will attempt the same contract call
	retryNum = 3
	// batchSize is the maximum number of upkeeps to register for each batch.
	batchSize             uint8 = 20
	upkeepIDLength        int   = 8
	conditionalUpkeepType uint8 = 0
	logtriggerUpkeepType  uint8 = 1
	DefaultRegisterAmount       = 10_000_000_000_000_000
	DefaultGasLimit             = 500_000
	DefaultCheckGas             = 10_000
	DefaultPerformGas           = 1_000
)

type VerifiableLoadInteractionConfig struct {
	RegisterUpkeepCount      uint8
	RegisteredUpkeepInterval uint32
	CancelBeforeRegister     bool
	SendLINKBeforeRegister   bool
}

type VerifiableLoadLogTriggerDeployable struct {
	contract  *verifiableLogTrigger.VerifiableLoadLogTriggerUpkeep
	registrar config.AutomationRegistrarV21Contract
	cCfg      *config.VerifiableLoadContract
}

func NewVerifiableLoadLogTriggerDeployable(
	registrar config.AutomationRegistrarV21Contract,
	conf *config.VerifiableLoadContract,
) (*VerifiableLoadLogTriggerDeployable, error) {
	if conf.LoadType != config.LogTriggerLoad {
		return nil, fmt.Errorf("%w: unvalid config for log trigger load contract", ErrConfiguration)
	}

	return &VerifiableLoadLogTriggerDeployable{
		registrar: registrar,
		cCfg:      conf,
	}, nil
}

func (d *VerifiableLoadLogTriggerDeployable) Connect(
	ctx context.Context,
	deployer *Deployer,
) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(d.cCfg.Address), deployer)
}

func (d *VerifiableLoadLogTriggerDeployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.registrar.Address)

	contractAddr, trx, _, err := verifiableLogTrigger.DeployVerifiableLoadLogTriggerUpkeep(
		opts, deployer.Client, registrar,
		d.cCfg.UseArbitrum, d.cCfg.UseMercury)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: Verifiable Load Contract creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	/*
		PrintVerifyContractCommand(
			config,
			contractAddr.String(),
			registrar.String(),
			fmt.Sprintf("%t", d.cCfg.UseArbitrum),
			fmt.Sprintf("%t", d.cCfg.UseMercury),
		)
	*/

	d.cCfg.Address = contractAddr.Hex()

	return contractAddr, nil
}

func (d *VerifiableLoadLogTriggerDeployable) ReadStats(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(d.cCfg.Address)

	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	// get all the stats from this block
	blockNum, err := deployer.Client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get block number: %s", ErrContractRead, err.Error())
	}

	opts := &bind.CallOpts{
		From:        deployer.Address,
		Context:     ctx,
		BlockNumber: big.NewInt(int64(blockNum)),
	}

	// get all active upkeep IDs on this verifiable load contract
	upkeepIds, err := contract.GetActiveUpkeepIDsDeployedByThisContract(opts, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return fmt.Errorf("%w: failed to get active upkeep IDs from %s: %s", ErrContractRead, addr, err.Error())
	}

	return collectAndWriteStats(ctx, contract, upkeepIds, blockNum, opts)
}

//nolint:cyclop
func (d *VerifiableLoadLogTriggerDeployable) RegisterUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(d.cCfg.Address)

	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if conf.CancelBeforeRegister {
		if err := cancelAllUpkeeps(ctx, contract, deployer); err != nil {
			return fmt.Errorf("%w: failed to cancel upkeeps: %s", ErrContractConnection, err.Error())
		}
	}

	if conf.SendLINKBeforeRegister {
		fmt.Printf("sending link to address: %s\n", d.cCfg.Address)
		// send the exact amount of LINK to the contract to run all deployed upkeeps
		amount := uint64(conf.RegisterUpkeepCount) * DefaultRegisterAmount * 2
		if err := deployer.SendLINK(ctx, d.cCfg.Address, amount); err != nil {
			return err
		}
	}

	upkeepIDs, err := registerNewUpkeeps(ctx, contract, deployer, conf.RegisterUpkeepCount, logtriggerUpkeepType)
	if err != nil {
		return fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	var offset int

	upkeepIDBatchSize := 25

	// batch calls have limited calldata size so do the upkeep ids in batches
	for offset < len(upkeepIDs) {
		var slice []*big.Int

		if offset+upkeepIDBatchSize > len(upkeepIDs) {
			slice = upkeepIDs[offset:]
		} else {
			slice = upkeepIDs[offset : offset+upkeepIDBatchSize]
		}

		if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			trx, err := contract.BatchSetIntervals(opts, slice, conf.RegisteredUpkeepInterval)
			if err != nil {
				return nil, fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
			}

			return trx, nil
		}); err != nil {
			return err
		}

		if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			trx, err := contract.BatchPreparingUpkeepsSimple(opts, slice, 0, 0)
			if err != nil {
				return nil, fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
			}

			return trx, nil
		}); err != nil {
			return err
		}

		offset += upkeepIDBatchSize
	}

	return runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		trx, err := contract.BatchSendLogs(opts, 0)
		if err != nil {
			return nil, fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
		}

		return trx, nil
	})
}

func (d *VerifiableLoadLogTriggerDeployable) CancelUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(d.cCfg.Address)

	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if err := cancelAllUpkeeps(ctx, contract, deployer); err != nil {
		return fmt.Errorf("%w: failed to cancel upkeeps: %s", ErrContractConnection, err.Error())
	}

	return nil
}

func (d *VerifiableLoadLogTriggerDeployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}

type VerifiableLoadConditionalDeployable struct {
	contract  *verifiableConditional.VerifiableLoadUpkeep
	registrar config.AutomationRegistrarV21Contract
	cCfg      *config.VerifiableLoadContract
}

func NewVerifiableLoadConditionalDeployable(
	registrar config.AutomationRegistrarV21Contract,
	conf *config.VerifiableLoadContract,
) (*VerifiableLoadConditionalDeployable, error) {
	if conf.LoadType != config.ConditionalLoad {
		return nil, fmt.Errorf("%w: invalid load type config", ErrConfiguration)
	}

	return &VerifiableLoadConditionalDeployable{
		registrar: registrar,
		cCfg:      conf,
	}, nil
}

func (d *VerifiableLoadConditionalDeployable) Connect(
	ctx context.Context,
	addr string,
	deployer *Deployer,
) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *VerifiableLoadConditionalDeployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.registrar.Address)

	contractAddr, trx, _, err := verifiableConditional.DeployVerifiableLoadUpkeep(
		opts, deployer.Client, registrar,
		d.cCfg.UseArbitrum,
	)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: Verifiable Load Contract creation failed: %s", ErrContractCreate, err.Error())
	}

	fmt.Println("waiting on transaction -- hash:", trx.Hash(), ", nonce:", trx.Nonce())

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	/*
		PrintVerifyContractCommand(
			config,
			contractAddr.String(),
			registrar.String(),
			fmt.Sprintf("%t", d.cCfg.UseArbitrum),
		)
	*/

	d.cCfg.Address = contractAddr.Hex()

	return contractAddr, nil
}

func (d *VerifiableLoadConditionalDeployable) ReadStats(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(d.cCfg.Address)

	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	// get all the stats from this block
	blockNum, err := deployer.Client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get block number: %s", ErrContractRead, err.Error())
	}

	opts := &bind.CallOpts{
		From:        deployer.Address,
		Context:     ctx,
		BlockNumber: big.NewInt(int64(blockNum)),
	}

	// get all active upkeep IDs on this verifiable load contract
	upkeepIds, err := contract.GetActiveUpkeepIDsDeployedByThisContract(opts, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return fmt.Errorf("%w: failed to get active upkeep IDs from %s: %s", ErrContractRead, addr, err.Error())
	}

	if len(upkeepIds) == 0 {
		return fmt.Errorf("%w: no upkeeps registered", ErrContractRead)
	}

	return collectAndWriteStats(ctx, contract, upkeepIds, blockNum, opts)
}

func (d *VerifiableLoadConditionalDeployable) RegisterUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(d.cCfg.Address)

	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if conf.CancelBeforeRegister {
		if err := cancelAllUpkeeps(ctx, contract, deployer); err != nil {
			return fmt.Errorf("%w: failed to cancel upkeeps: %s", ErrContractConnection, err.Error())
		}
	}

	if conf.SendLINKBeforeRegister {
		// send the exact amount of LINK to the contract to run all deployed upkeeps
		amount := uint64(conf.RegisterUpkeepCount) * DefaultRegisterAmount

		if err := deployer.SendLINK(ctx, d.cCfg.Address, amount); err != nil {
			return err
		}
	}

	upkeepIDs, err := registerNewUpkeeps(ctx, contract, deployer, conf.RegisterUpkeepCount, conditionalUpkeepType)
	if err != nil {
		return fmt.Errorf("%w: failed to register upkeeps: %s", ErrContractConnection, err.Error())
	}

	if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		trx, err := contract.BatchSetIntervals(opts, upkeepIDs, conf.RegisteredUpkeepInterval)
		if err != nil {
			return nil, fmt.Errorf("%w: transaction failed %s", ErrContractConnection, err.Error())
		}

		return trx, nil
	}); err != nil {
		return err
	}

	return runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		trx, err := contract.BatchUpdatePipelineData(opts, upkeepIDs)
		if err != nil {
			return nil, fmt.Errorf("%w: transaction failed %s", ErrContractConnection, err.Error())
		}

		return trx, nil
	})
}

func (d *VerifiableLoadConditionalDeployable) CancelUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(d.cCfg.Address)

	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if err := cancelAllUpkeeps(ctx, contract, deployer); err != nil {
		return fmt.Errorf("%w: failed to cancel upkeeps: %s", ErrContractConnection, err.Error())
	}

	return nil
}

func (d *VerifiableLoadConditionalDeployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}

type upkeepInfo struct {
	mu              sync.Mutex
	ID              *big.Int
	Bucket          uint16
	DelayBuckets    map[uint16][]float64
	SortedAllDelays []float64
	TotalDelayBlock float64
	TotalPerforms   uint64
}

func (ui *upkeepInfo) AddBucket(bucketNum uint16, bucketDelays []float64) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.DelayBuckets[bucketNum] = bucketDelays
}

type upkeepStats struct {
	BlockNumber     uint64
	AllInfos        []*upkeepInfo
	TotalDelayBlock float64
	TotalPerforms   uint64
	SortedAllDelays []float64
}

type verifiableLoadContract interface {
	Counters(*bind.CallOpts, *big.Int) (*big.Int, error)
	Buckets(*bind.CallOpts, *big.Int) (uint16, error)
	GetBucketedDelays(*bind.CallOpts, *big.Int, uint16) ([]*big.Int, error)
}

func getUpkeepInfoJob(upkeepID *big.Int, contract verifiableLoadContract, opts *bind.CallOpts) util.Job[*upkeepInfo] {
	return func(ctx context.Context) *upkeepInfo {
		// fetch how many times this upkeep has been executed
		counter, err := contract.Counters(opts, upkeepID)
		if err != nil {
			log.Fatalf("failed to get counter for %s: %v", upkeepID.String(), err)
		}

		// get all the buckets of an upkeep. 100 performs is a bucket.
		bucket, err := contract.Buckets(opts, upkeepID)
		if err != nil {
			log.Fatalf("failed to get current bucket count for %s: %v", upkeepID.String(), err)
		}

		info := &upkeepInfo{
			ID:            upkeepID,
			Bucket:        bucket,
			TotalPerforms: counter.Uint64(),
			DelayBuckets:  map[uint16][]float64{},
		}

		var (
			delays []float64
			group  sync.WaitGroup
		)

		for idx := uint16(0); idx <= bucket; idx++ {
			group.Add(1)

			go func(idx uint16) {
				defer group.Done()

				getBucketData(contract, opts, upkeepID, idx, info)
			}(idx)
		}

		group.Wait()

		for i := uint16(0); i <= bucket; i++ {
			bucketDelays := info.DelayBuckets[i]
			delays = append(delays, bucketDelays...)

			for _, d := range bucketDelays {
				info.TotalDelayBlock += d
			}
		}

		sort.Float64s(delays)
		info.SortedAllDelays = delays
		info.TotalPerforms = uint64(len(info.SortedAllDelays))

		return info
	}
}

//nolint:funlen
func getUpkeepInfo(
	writer table.Writer,
	idChan chan *big.Int,
	resultsChan chan *upkeepInfo,
	contract verifiableLoadContract,
	opts *bind.CallOpts,
) {
	for upkeepID := range idChan {
		// fetch how many times this upkeep has been executed
		counter, err := contract.Counters(opts, upkeepID)
		if err != nil {
			log.Fatalf("failed to get counter for %s: %v", upkeepID.String(), err)
		}

		// get all the buckets of an upkeep. 100 performs is a bucket.
		bucket, err := contract.Buckets(opts, upkeepID)
		if err != nil {
			log.Fatalf("failed to get current bucket count for %s: %v", upkeepID.String(), err)
		}

		info := &upkeepInfo{
			ID:            upkeepID,
			Bucket:        bucket,
			TotalPerforms: counter.Uint64(),
			DelayBuckets:  map[uint16][]float64{},
		}

		var (
			delays []float64
			group  sync.WaitGroup
		)

		for idx := uint16(0); idx <= bucket; idx++ {
			group.Add(1)

			go func(idx uint16) {
				defer group.Done()

				getBucketData(contract, opts, upkeepID, idx, info)
			}(idx)
		}

		group.Wait()

		for i := uint16(0); i <= bucket; i++ {
			bucketDelays := info.DelayBuckets[i]
			delays = append(delays, bucketDelays...)

			for _, d := range bucketDelays {
				info.TotalDelayBlock += d
			}
		}

		sort.Float64s(delays)
		info.SortedAllDelays = delays
		info.TotalPerforms = uint64(len(info.SortedAllDelays))

		var maxDelay float64

		percentiles := make([]float64, percentilesCalculated)

		if len(info.SortedAllDelays) > 0 {
			percentiles, err = getPercentiles(info.SortedAllDelays, P50, P90, P95, P99)
			if err != nil {
				log.Println(err)
			}

			maxDelay = info.SortedAllDelays[len(info.SortedAllDelays)-1]
		}

		writer.AppendRow(table.Row{
			shorten(upkeepID.String(), upkeepIDLength),
			fmt.Sprintf("%d", info.TotalPerforms),
			fmt.Sprintf("%f", percentiles[0]),
			fmt.Sprintf("%f", percentiles[1]),
			fmt.Sprintf("%f", percentiles[2]),
			fmt.Sprintf("%f", percentiles[3]),
			fmt.Sprintf("%f", maxDelay),
			fmt.Sprintf("%d", uint64(info.TotalDelayBlock)),
			fmt.Sprintf("%f", info.TotalDelayBlock/float64(info.TotalPerforms)),
		}, table.RowConfig{AutoMerge: true})

		resultsChan <- info
	}
}

func getBucketData(
	contract verifiableLoadContract,
	opts *bind.CallOpts,
	upkeepID *big.Int,
	bucketNum uint16,
	info *upkeepInfo,
) {
	var (
		bucketDelays []*big.Int
		err          error
	)

	for i := 0; i < retryNum; i++ {
		bucketDelays, err = contract.GetBucketedDelays(opts, upkeepID, bucketNum)
		if err != nil {
			log.Printf(
				"failed to get bucketed delays for upkeep id %s bucket %d: %v, retrying...",
				upkeepID.String(),
				bucketNum,
				err,
			)

			time.Sleep(retryDelay)
		} else {
			break
		}
	}

	floatBucketDelays := make([]float64, 0, len(bucketDelays))

	for _, d := range bucketDelays {
		floatBucketDelays = append(floatBucketDelays, float64(d.Uint64()))
	}

	sort.Float64s(floatBucketDelays)
	info.AddBucket(bucketNum, floatBucketDelays)
}

func shorten(full string, outLen int) string {
	if utf8.RuneCountInString(full) < outLen {
		return full
	}

	return string([]byte(full)[:outLen])
}

func getPercentiles(delays []float64, percentiles ...float64) ([]float64, error) {
	calculated := make([]float64, len(percentiles))

	for idx, percentile := range percentiles {
		var err error

		calculated[idx], err = stats.Percentile(delays, percentile)

		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCalculationFailure, err.Error())
		}
	}

	return calculated, nil
}

type upkeepCanceller interface {
	GetActiveUpkeepIDsDeployedByThisContract(*bind.CallOpts, *big.Int, *big.Int) ([]*big.Int, error)
	BatchCancelUpkeeps(*bind.TransactOpts, []*big.Int) (*types.Transaction, error)
}

func cancelAllUpkeeps(ctx context.Context, canceller upkeepCanceller, deployer *Deployer) error {
	// there is likely a limit on the number of upkeeps this function can return
	// TODO: do this in batches
	oldUpkeepIds, err := canceller.GetActiveUpkeepIDsDeployedByThisContract(
		&bind.CallOpts{
			Context: ctx,
			From:    deployer.Address,
		},
		big.NewInt(0),
		big.NewInt(0),
	)
	if err != nil {
		return fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	if len(oldUpkeepIds) == 0 {
		return nil
	}

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	// TODO: do this in batches for larger number of upkeeps
	trx, err := canceller.BatchCancelUpkeeps(opts, oldUpkeepIds)
	if err != nil {
		return fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	if err := deployer.wait(ctx, trx); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	return nil
}

type upkeepRegister interface {
	//nolint:lll
	BatchRegisterUpkeeps(*bind.TransactOpts, uint8, uint32, uint8, []byte, *big.Int, *big.Int, *big.Int) (*types.Transaction, error)
	GetActiveUpkeepIDsDeployedByThisContract(*bind.CallOpts, *big.Int, *big.Int) ([]*big.Int, error)
}

func registerNewUpkeeps(
	ctx context.Context,
	register upkeepRegister,
	deployer *Deployer,
	count uint8,
	typ uint8,
) ([]*big.Int, error) {
	var completed uint8 = 0

	for completed < count {
		setSize := count - completed
		if setSize > batchSize {
			setSize = batchSize
		}

		opts, err := deployer.BuildTxOpts(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
		}

		trx, err := register.BatchRegisterUpkeeps(
			opts, setSize, DefaultGasLimit,
			typ, []byte{0x00},
			big.NewInt(DefaultRegisterAmount),
			big.NewInt(DefaultCheckGas),
			big.NewInt(DefaultPerformGas),
		)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to register upkeeps: %s", ErrContractConnection, err.Error())
		}

		if err := deployer.wait(ctx, trx); err != nil {
			return nil, fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
		}

		completed = completed + setSize
	}

	upkeepOpts := &bind.CallOpts{
		Context: ctx,
		From:    deployer.Address,
	}

	upkeepIDs, err := register.GetActiveUpkeepIDsDeployedByThisContract(upkeepOpts, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	return upkeepIDs, nil
}

func runContractFunc(
	ctx context.Context,
	deployer *Deployer,
	contractFn func(*bind.TransactOpts) (*types.Transaction, error),
) error {
	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	trx, err := contractFn(opts)
	if err != nil {
		return fmt.Errorf("%w: failed to set upkeep intervals: %s", ErrContractConnection, err.Error())
	}

	if err := deployer.wait(ctx, trx); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	return nil
}

//nolint:funlen
func collectAndWriteStats(
	ctx context.Context,
	contract verifiableLoadContract,
	upkeepIDs []*big.Int,
	block uint64,
	opts *bind.CallOpts,
) error {
	uStats := &upkeepStats{BlockNumber: block}
	jobs := make([]util.Job[*upkeepInfo], len(upkeepIDs))

	for idx := range upkeepIDs {
		jobs[idx] = getUpkeepInfoJob(upkeepIDs[idx], contract, opts)
	}

	results := util.NewParallel[*upkeepInfo](workerNum).RunWithContext(ctx, jobs)
	writer := table.NewWriter()

	writer.SetTitle(fmt.Sprintf("Upkeep Results --> (All STATS BELOW ARE CALCULATED AT BLOCK %d)", block))
	writer.AppendHeader(headerRowOne, table.RowConfig{AutoMerge: true})
	writer.AppendHeader(headerRowTwo)

	for _, info := range results {
		var (
			maxDelay float64
			err      error
		)

		percentiles := make([]float64, percentilesCalculated)

		if len(info.SortedAllDelays) > 0 {
			percentiles, err = getPercentiles(info.SortedAllDelays, P50, P90, P95, P99)
			if err != nil {
				return err
			}

			maxDelay = info.SortedAllDelays[len(info.SortedAllDelays)-1]
		}

		writer.AppendRow(table.Row{
			shorten(info.ID.String(), upkeepIDLength),
			fmt.Sprintf("%d", info.TotalPerforms),
			fmt.Sprintf("%f", percentiles[0]),
			fmt.Sprintf("%f", percentiles[1]),
			fmt.Sprintf("%f", percentiles[2]),
			fmt.Sprintf("%f", percentiles[3]),
			fmt.Sprintf("%f", maxDelay),
			fmt.Sprintf("%d", uint64(info.TotalDelayBlock)),
			fmt.Sprintf("%f", info.TotalDelayBlock/float64(info.TotalPerforms)),
		}, table.RowConfig{AutoMerge: true})

		uStats.AllInfos = append(uStats.AllInfos, info)
		uStats.TotalPerforms += info.TotalPerforms
		uStats.TotalDelayBlock += info.TotalDelayBlock
		uStats.SortedAllDelays = append(uStats.SortedAllDelays, info.SortedAllDelays...)
	}

	sort.Float64s(uStats.SortedAllDelays)

	var (
		maxDelay float64
		err      error
	)

	percentiles := make([]float64, percentilesCalculated)

	if len(uStats.SortedAllDelays) > 0 {
		percentiles, err = getPercentiles(uStats.SortedAllDelays, P50, P90, P95, P99)
		if err != nil {
			return fmt.Errorf("%w: percentile calculation: %s", ErrContractRead, err.Error())
		}

		maxDelay = uStats.SortedAllDelays[len(uStats.SortedAllDelays)-1]
	}

	writer.AppendFooter(table.Row{
		"Total",
		uStats.TotalPerforms,
		fmt.Sprintf("%f", percentiles[0]),
		fmt.Sprintf("%f", percentiles[1]),
		fmt.Sprintf("%f", percentiles[2]),
		fmt.Sprintf("%f", percentiles[3]),
		fmt.Sprintf("%f", maxDelay),
		fmt.Sprintf("%f", uStats.TotalDelayBlock),
		fmt.Sprintf("%f", uStats.TotalDelayBlock/float64(uStats.TotalPerforms)),
	})

	writer.SetAutoIndex(true)
	writer.SetStyle(table.StyleLight)

	writer.Style().Options.SeparateRows = true

	//nolint:forbidigo
	fmt.Println(writer.Render())

	return nil
}
