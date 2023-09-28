package asset

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/montanaflynn/stats"

	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_log_trigger_upkeep_wrapper"
	verifiableLogTrigger "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_log_trigger_upkeep_wrapper"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_upkeep_wrapper"
	verifiableConditional "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_upkeep_wrapper"
)

type VerifiableLoadConfig struct {
	RegistrarAddr string
	UseMercury    bool
	UseArbitrum   bool
	AutoLog       bool
}

type VerifiableLoadInteractionConfig struct {
	ContractAddr string
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
	config VerifyContractConfig,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.cCfg.RegistrarAddr)

	contractAddr, trx, _, err := verifiableLogTrigger.DeployVerifiableLoadLogTriggerUpkeep(
		opts, deployer.Client, registrar,
		d.cCfg.UseArbitrum, d.cCfg.AutoLog, d.cCfg.UseMercury)
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
			fmt.Sprintf("%t", d.cCfg.AutoLog),
			fmt.Sprintf("%t", d.cCfg.UseMercury),
		)
	*/

	return contractAddr, nil
}

func (d *VerifiableLoadLogTriggerDeployable) ReadStats(ctx context.Context, deployer *Deployer, conf VerifiableLoadInteractionConfig) error {
	addr := common.HexToAddress(conf.ContractAddr)

	contract, err := verifiable_load_log_trigger_upkeep_wrapper.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	// get all the stats from this block
	blockNum, err := deployer.Client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number: %s", err.Error())
	}

	opts := &bind.CallOpts{
		From:        deployer.Address,
		Context:     ctx,
		BlockNumber: big.NewInt(int64(blockNum)),
	}

	// get all active upkeep IDs on this verifiable load contract
	upkeepIds, err := contract.GetActiveUpkeepIDs(opts, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return fmt.Errorf("failed to get active upkeep IDs from %s: %s", addr, err.Error())
	}

	us := &upkeepStats{BlockNumber: blockNum}

	resultsChan := make(chan *upkeepInfo, maxUpkeepNum)
	idChan := make(chan *big.Int, maxUpkeepNum)

	var wg sync.WaitGroup

	// create a number of workers to process the upkeep ids in batch
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go getUpkeepInfo(idChan, resultsChan, contract, opts, &wg)
	}

	for _, id := range upkeepIds {
		idChan <- id
	}

	close(idChan)
	wg.Wait()

	close(resultsChan)

	for info := range resultsChan {
		us.AllInfos = append(us.AllInfos, info)
		us.TotalPerforms += info.TotalPerforms
		us.TotalDelayBlock += info.TotalDelayBlock
		us.SortedAllDelays = append(us.SortedAllDelays, info.SortedAllDelays...)
	}

	sort.Float64s(us.SortedAllDelays)

	log.Println("\n\n================================== ALL UPKEEPS SUMMARY =======================================================")
	p50, _ := stats.Percentile(us.SortedAllDelays, 50)
	p90, _ := stats.Percentile(us.SortedAllDelays, 90)
	p95, _ := stats.Percentile(us.SortedAllDelays, 95)
	p99, _ := stats.Percentile(us.SortedAllDelays, 99)
	maxDelay := us.SortedAllDelays[len(us.SortedAllDelays)-1]
	log.Printf("For total %d upkeeps: total performs: %d, p50: %f, p90: %f, p95: %f, p99: %f, max delay: %f, total delay blocks: %f, average perform delay: %f\n", len(upkeepIds), us.TotalPerforms, p50, p90, p95, p99, maxDelay, us.TotalDelayBlock, us.TotalDelayBlock/float64(us.TotalPerforms))
	log.Printf("All STATS ABOVE ARE CALCULATED AT BLOCK %d", blockNum)

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

	contractAddr, trx, _, err := verifiableConditional.DeployVerifiableLoadUpkeep(
		opts, deployer.Client, registrar,
		d.cCfg.UseArbitrum,
	)
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
		)
	*/

	return contractAddr, nil
}

const (
	// workerNum is the total number of workers calculating upkeeps' delay summary
	workerNum = 5
	// retryDelay is the time the go routine will wait before calling the same contract function
	retryDelay = 1 * time.Second
	// retryNum defines how many times the go routine will attempt the same contract call
	retryNum = 3
	// maxUpkeepNum defines the size of channels. Increase if there are lots of upkeeps.
	maxUpkeepNum = 100
)

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

func (d *VerifiableLoadConditionalDeployable) ReadStats(ctx context.Context, deployer *Deployer, conf VerifiableLoadInteractionConfig) error {
	addr := common.HexToAddress(conf.ContractAddr)

	contract, err := verifiable_load_upkeep_wrapper.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return fmt.Errorf("failed to create a new verifiable load upkeep from address %s: %v", addr, err)
	}

	// get all the stats from this block
	blockNum, err := deployer.Client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block number: %s", err.Error())
	}

	opts := &bind.CallOpts{
		From:        deployer.Address,
		Context:     ctx,
		BlockNumber: big.NewInt(int64(blockNum)),
	}

	// get all active upkeep IDs on this verifiable load contract
	upkeepIds, err := contract.GetActiveUpkeepIDs(opts, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return fmt.Errorf("failed to get active upkeep IDs from %s: %s", addr, err.Error())
	}

	us := &upkeepStats{BlockNumber: blockNum}

	resultsChan := make(chan *upkeepInfo, maxUpkeepNum)
	idChan := make(chan *big.Int, maxUpkeepNum)

	var wg sync.WaitGroup

	// create a number of workers to process the upkeep ids in batch
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go getUpkeepInfo(idChan, resultsChan, contract, opts, &wg)
	}

	for _, id := range upkeepIds {
		idChan <- id
	}

	close(idChan)
	wg.Wait()

	close(resultsChan)

	for info := range resultsChan {
		us.AllInfos = append(us.AllInfos, info)
		us.TotalPerforms += info.TotalPerforms
		us.TotalDelayBlock += info.TotalDelayBlock
		us.SortedAllDelays = append(us.SortedAllDelays, info.SortedAllDelays...)
	}

	sort.Float64s(us.SortedAllDelays)

	log.Println("\n\n================================== ALL UPKEEPS SUMMARY =======================================================")
	p50, _ := stats.Percentile(us.SortedAllDelays, 50)
	p90, _ := stats.Percentile(us.SortedAllDelays, 90)
	p95, _ := stats.Percentile(us.SortedAllDelays, 95)
	p99, _ := stats.Percentile(us.SortedAllDelays, 99)
	maxDelay := us.SortedAllDelays[len(us.SortedAllDelays)-1]
	log.Printf("For total %d upkeeps: total performs: %d, p50: %f, p90: %f, p95: %f, p99: %f, max delay: %f, total delay blocks: %f, average perform delay: %f\n", len(upkeepIds), us.TotalPerforms, p50, p90, p95, p99, maxDelay, us.TotalDelayBlock, us.TotalDelayBlock/float64(us.TotalPerforms))
	log.Printf("All STATS ABOVE ARE CALCULATED AT BLOCK %d", blockNum)

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

type verifiableLoadContract interface {
	Counters(*bind.CallOpts, *big.Int) (*big.Int, error)
	Buckets(*bind.CallOpts, *big.Int) (uint16, error)
	GetBucketedDelays(*bind.CallOpts, *big.Int, uint16) ([]*big.Int, error)
}

func getUpkeepInfo(idChan chan *big.Int, resultsChan chan *upkeepInfo, contract verifiableLoadContract, opts *bind.CallOpts, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range idChan {
		// fetch how many times this upkeep has been executed
		c, err := contract.Counters(opts, id)
		if err != nil {
			log.Fatalf("failed to get counter for %s: %v", id.String(), err)
		}

		// get all the buckets of an upkeep. 100 performs is a bucket.
		b, err := contract.Buckets(opts, id)
		if err != nil {
			log.Fatalf("failed to get current bucket count for %s: %v", id.String(), err)
		}

		info := &upkeepInfo{
			ID:            id,
			Bucket:        b,
			TotalPerforms: c.Uint64(),
			DelayBuckets:  map[uint16][]float64{},
		}

		var delays []float64
		var wg1 sync.WaitGroup
		for i := uint16(0); i <= b; i++ {
			wg1.Add(1)
			go getBucketData(contract, opts, id, i, &wg1, info)
		}
		wg1.Wait()

		for i := uint16(0); i <= b; i++ {
			bucketDelays := info.DelayBuckets[i]
			delays = append(delays, bucketDelays...)
			for _, d := range bucketDelays {
				info.TotalDelayBlock += d
			}
		}
		sort.Float64s(delays)
		info.SortedAllDelays = delays
		info.TotalPerforms = uint64(len(info.SortedAllDelays))

		p50, _ := stats.Percentile(info.SortedAllDelays, 50)
		p90, _ := stats.Percentile(info.SortedAllDelays, 90)
		p95, _ := stats.Percentile(info.SortedAllDelays, 95)
		p99, _ := stats.Percentile(info.SortedAllDelays, 99)
		maxDelay := info.SortedAllDelays[len(info.SortedAllDelays)-1]

		log.Printf("upkeep ID %s has %d performs in total. p50: %f, p90: %f, p95: %f, p99: %f, max delay: %f, total delay blocks: %d, average perform delay: %f\n", id, info.TotalPerforms, p50, p90, p95, p99, maxDelay, uint64(info.TotalDelayBlock), info.TotalDelayBlock/float64(info.TotalPerforms))
		resultsChan <- info
	}
}

func getBucketData(contract verifiableLoadContract, opts *bind.CallOpts, id *big.Int, bucketNum uint16, wg *sync.WaitGroup, info *upkeepInfo) {
	defer wg.Done()

	var bucketDelays []*big.Int
	var err error
	for i := 0; i < retryNum; i++ {
		bucketDelays, err = contract.GetBucketedDelays(opts, id, bucketNum)
		if err != nil {
			log.Printf("failed to get bucketed delays for upkeep id %s bucket %d: %v, retrying...", id.String(), bucketNum, err)
			time.Sleep(retryDelay)
		} else {
			break
		}
	}

	var floatBucketDelays []float64
	for _, d := range bucketDelays {
		floatBucketDelays = append(floatBucketDelays, float64(d.Uint64()))
	}
	sort.Float64s(floatBucketDelays)
	info.AddBucket(bucketNum, floatBucketDelays)
}
