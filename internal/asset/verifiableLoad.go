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
	ErrContractRead = fmt.Errorf("contract read error")
	//nolint:gochecknoglobals
	headerRowOne = table.Row{
		"ID", "Total Performs", "Block Delays", "Block Delays", "Block Delays",
		"Block Delays", "Block Delays", "Block Delays", "Block Delays"}
	//nolint:gochecknoglobals
	headerRowTwo = table.Row{
		"", "", "50th", "90th", "95th", "99th", "Max", "Total", "Average"}
)

const (
	P50 float64 = 50
	P90 float64 = 90
	P95 float64 = 95
	P99 float64 = 99
	// workerNum is the total number of workers calculating upkeeps' delay summary
	workerNum = 5
	// retryDelay is the time the go routine will wait before calling the same contract function
	retryDelay = 1 * time.Second
	// retryNum defines how many times the go routine will attempt the same contract call
	retryNum = 3
	// maxUpkeepNum defines the size of channels. Increase if there are lots of upkeeps.
	maxUpkeepNum                = 100
	upkeepIDLength        int   = 8
	conditionalUpkeepType uint8 = 0
	logtriggerUpkeepType  uint8 = 1
	DefaultRegisterAmount       = 1_000_000_000_000_000_000
	DefaultGasLimit             = 500_000
	DefaultCheckGas             = 10_000
	DefaultPerformGas           = 1_000
)

type VerifiableLoadConfig struct {
	RegistrarAddr string
	UseMercury    bool
	UseArbitrum   bool
}

type VerifiableLoadInteractionConfig struct {
	ContractAddr             string
	RegisterUpkeepCount      uint8
	RegisteredUpkeepInterval uint32
	CancelBeforeRegister     bool
}

type VerifiableLoadLogTriggerDeployable struct {
	contract *verifiableLogTrigger.VerifiableLoadLogTriggerUpkeep
	cCfg     *VerifiableLoadConfig
}

func NewVerifiableLoadLogTriggerDeployable(cCfg *VerifiableLoadConfig) *VerifiableLoadLogTriggerDeployable {
	return &VerifiableLoadLogTriggerDeployable{
		cCfg: cCfg,
	}
}

func (d *VerifiableLoadLogTriggerDeployable) Connect(
	ctx context.Context,
	addr string,
	deployer *Deployer,
) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *VerifiableLoadLogTriggerDeployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
	_ VerifyContractConfig,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.cCfg.RegistrarAddr)

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

	return contractAddr, nil
}

//nolint:funlen
func (d *VerifiableLoadLogTriggerDeployable) ReadStats(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(conf.ContractAddr)

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

func (d *VerifiableLoadLogTriggerDeployable) RegisterUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(conf.ContractAddr)

	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if conf.CancelBeforeRegister {
		if err := cancelUpkeeps(ctx, contract, deployer, int64(conf.RegisterUpkeepCount)); err != nil {
			return fmt.Errorf("%w: failed to cancel upkeeps: %s", ErrContractConnection, err.Error())
		}
	}

	upkeepIDs, err := registerNewUpkeeps(ctx, contract, deployer, int64(conf.RegisterUpkeepCount), logtriggerUpkeepType)
	if err != nil {
		return fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.BatchSetIntervals(opts, upkeepIDs, conf.RegisteredUpkeepInterval)
	}); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.BatchPreparingUpkeepsSimple(opts, upkeepIDs, 0, 0)
	}); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.BatchSendLogs(opts, 0)
	}); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	return nil
}

func (d *VerifiableLoadLogTriggerDeployable) CancelUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(conf.ContractAddr)

	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if err := cancelUpkeeps(ctx, contract, deployer, int64(conf.RegisterUpkeepCount)); err != nil {
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
	contract *verifiableConditional.VerifiableLoadUpkeep
	cCfg     *VerifiableLoadConfig
}

func NewVerifiableLoadConditionalDeployable(cCfg *VerifiableLoadConfig) *VerifiableLoadConditionalDeployable {
	return &VerifiableLoadConditionalDeployable{
		cCfg: cCfg,
	}
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
	config VerifyContractConfig,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.cCfg.RegistrarAddr)

	// fmt.Printf("%+v\n", opts)
	// return registrar, fmt.Errorf("stop")

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

	return contractAddr, nil
}

//nolint:funlen
func (d *VerifiableLoadConditionalDeployable) ReadStats(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(conf.ContractAddr)

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

//nolint:funlen,cyclop
func (d *VerifiableLoadConditionalDeployable) RegisterUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(conf.ContractAddr)

	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if conf.CancelBeforeRegister {
		if err := cancelUpkeeps(ctx, contract, deployer, int64(conf.RegisterUpkeepCount)); err != nil {
			return fmt.Errorf("%w: failed to cancel upkeeps: %s", ErrContractConnection, err.Error())
		}
	}

	upkeepIDs, err := registerNewUpkeeps(ctx, contract, deployer, int64(conf.RegisterUpkeepCount), conditionalUpkeepType)
	if err != nil {
		return fmt.Errorf("%w: failed to register upkeeps: %s", ErrContractConnection, err.Error())
	}

	if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.BatchSetIntervals(opts, upkeepIDs, conf.RegisteredUpkeepInterval)
	}); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	if err := runContractFunc(ctx, deployer, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return contract.BatchUpdatePipelineData(opts, upkeepIDs)
	}); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	return nil
}

func (d *VerifiableLoadConditionalDeployable) CancelUpkeeps(
	ctx context.Context,
	deployer *Deployer,
	conf VerifiableLoadInteractionConfig,
) error {
	addr := common.HexToAddress(conf.ContractAddr)

	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	if err := cancelUpkeeps(ctx, contract, deployer, int64(conf.RegisterUpkeepCount)); err != nil {
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

		percentiles := make([]float64, 4)

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
			return nil, err
		}
	}

	return calculated, nil
}

type upkeepCanceller interface {
	GetActiveUpkeepIDsDeployedByThisContract(*bind.CallOpts, *big.Int, *big.Int) ([]*big.Int, error)
	BatchCancelUpkeeps(*bind.TransactOpts, []*big.Int) (*types.Transaction, error)
}

func cancelUpkeeps(ctx context.Context, canceller upkeepCanceller, deployer *Deployer, count int64) error {
	oldUpkeepIds, err := canceller.GetActiveUpkeepIDsDeployedByThisContract(
		&bind.CallOpts{
			Context: ctx,
			From:    deployer.Address,
		},
		big.NewInt(0),
		big.NewInt(count),
	)
	if err != nil {
		return fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	trx, err := canceller.BatchCancelUpkeeps(nil, oldUpkeepIds)
	if err != nil {
		return fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	if err := deployer.wait(ctx, trx); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	return nil
}

type upkeepRegister interface {
	BatchRegisterUpkeeps(*bind.TransactOpts, uint8, uint32, uint8, []byte, *big.Int, *big.Int, *big.Int) (*types.Transaction, error)
	GetActiveUpkeepIDsDeployedByThisContract(*bind.CallOpts, *big.Int, *big.Int) ([]*big.Int, error)
}

func registerNewUpkeeps(ctx context.Context, register upkeepRegister, deployer *Deployer, count int64, typ uint8) ([]*big.Int, error) {
	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	trx, err := register.BatchRegisterUpkeeps(
		opts, uint8(count), DefaultGasLimit,
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

	upkeepOpts := &bind.CallOpts{
		Context: ctx,
		From:    deployer.Address,
	}

	upkeepIDs, err := register.GetActiveUpkeepIDsDeployedByThisContract(upkeepOpts, big.NewInt(0), big.NewInt(count))
	if err != nil {
		return nil, fmt.Errorf("%w: contract query failed: %s", ErrContractConnection, err.Error())
	}

	return upkeepIDs, nil
}

func runContractFunc(ctx context.Context, deployer *Deployer, fn func(*bind.TransactOpts) (*types.Transaction, error)) error {
	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	trx, err := fn(opts)
	if err != nil {
		return fmt.Errorf("%w: failed to set upkeep intervals: %s", ErrContractConnection, err.Error())
	}

	if err := deployer.wait(ctx, trx); err != nil {
		return fmt.Errorf("%w: transaction failed: %s", ErrContractConnection, err.Error())
	}

	return nil
}

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

	results := util.NewParallel[*upkeepInfo](20).RunWithContext(ctx, jobs)
	writer := table.NewWriter()

	writer.SetTitle(fmt.Sprintf("Upkeep Results --> (All STATS BELOW ARE CALCULATED AT BLOCK %d)", block))
	writer.AppendHeader(headerRowOne, table.RowConfig{AutoMerge: true})
	writer.AppendHeader(headerRowTwo)

	for _, info := range results {
		var maxDelay float64
		var err error

		percentiles := make([]float64, 4)

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

	var maxDelay float64
	var err error

	percentiles := make([]float64, 4)

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
