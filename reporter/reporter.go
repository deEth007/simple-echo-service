package reporter

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// L1 Gas는 L2에서 지출한 것만 기록 즉, Withdrawal 트랜잭션에서 발생한 L1가스 fee 기록

// Deposit
// L2StandardBridge : 0x4200000000000000000000000000000000000010
// Event : DepositFinalized(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
// 해당 Event를 가진 트랜잭션을 검색할 수 있나? -> 그러면 L2에서 사용된 Gas를 알 수 있음 (최종 결과 합산만) -> 가능!

// Withdrawal
// 총 출금 개수는 MessagePassed 개수를 기록하면 됨
// L2ToL1MessagePasser: 0x4200000000000000000000000000000000000016
// Event: MessagePassed(uint256 indexed nonce, address indexed sender, address indexed target, uint256 value, uint256 gasLimit, bytes data, bytes32 withdrawalHash)
// MessagePassed 이벤트가 발생하면 L2 측에서는 출금에 성공한 것이 됨
// 그냥 컨펌 기다리면 됨

// 총 batch 제출 수? -> batcher 계정 조하면 이더랑 제출 수 조회 가능
// 총 proposer "                "

// SequencerFeeVault 0x4200000000000000000000000000000000000011

type reports struct {
	tps                       *big.Int
	totalConfirmTransactions  *big.Int
	l1GasUsed                 *big.Int
	l2GasUsed                 *big.Int
	l1GasFee                  *big.Int
	l2GasFee                  *big.Int
	batcherConsumeEther       *big.Int
	proposerConsumeEther      *big.Int
	totalSequncerConsumeEther *big.Int
	startBlockNumber          *big.Int
	endBlockNumber            *big.Int
	l2BlockTime               *big.Int
}

var once sync.Once

var reporter *reports

func (r *reports) RecordTPS() {
	sb := new(big.Int).Set(r.startBlockNumber)
	eb := new(big.Int).Set(r.endBlockNumber)
	tb := eb.Sub(eb, sb)
	duration := tb.Mul(tb, r.l2BlockTime)

	tr := new(big.Int).Set(r.totalConfirmTransactions)
	if duration.Cmp(big.NewInt(0)) != 0 {
		r.tps = tr.Div(tr, duration)
	}
}

func (r *reports) RecordConfirmRequest() {
	r.totalConfirmTransactions.Add(r.totalConfirmTransactions, big.NewInt(1))
}

// 수정해야 됨 start 항상 0임
func (r *reports) RecordStartToLastBlock(receipt *types.Receipt) {
	if r.startBlockNumber.Cmp(receipt.BlockNumber) > 0 {
		r.startBlockNumber.Set(receipt.BlockNumber)
	}
	if r.endBlockNumber.Cmp(receipt.BlockNumber) < 0 {
		r.endBlockNumber.Set(receipt.BlockNumber)
	}
}

func (r *reports) RecordL1GasUsed(receipt *types.Receipt) {
	r.l1GasUsed.Add(r.l1GasUsed, receipt.L1GasUsed)
}

func (r *reports) RecordL2GasUsed(receipt *types.Receipt) {
	r.l2GasUsed.Add(r.l2GasUsed, new(big.Int).SetUint64(receipt.GasUsed))
}

func (r *reports) RecordL1GasFee(receipt *types.Receipt) {
	r.l1GasFee.Add(r.l1GasFee, receipt.L1Fee)
}

func (r *reports) RecordL2GasFee(receipt *types.Receipt) {
	l2GasFee := receipt.EffectiveGasPrice.Mul(receipt.EffectiveGasPrice, new(big.Int).SetUint64(receipt.GasUsed))
	r.l2GasFee.Add(r.l2GasFee, l2GasFee)
}

func weiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float).SetInt(wei)
	wetiToEtherFactor := new(big.Float).SetInt(big.NewInt(params.Ether))
	ether.Quo(ether, wetiToEtherFactor)
	return ether
}

func Get() *reports {
	return reporter
}

func InitReporter(cfg *Config) {
	once.Do(
		func() {
			reporter = &reports{
				tps:                       big.NewInt(0),
				totalConfirmTransactions:  big.NewInt(0),
				l1GasUsed:                 big.NewInt(0),
				l2GasUsed:                 big.NewInt(0),
				l1GasFee:                  big.NewInt(0),
				l2GasFee:                  big.NewInt(0),
				batcherConsumeEther:       big.NewInt(0),
				proposerConsumeEther:      big.NewInt(0),
				startBlockNumber:          big.NewInt(0),
				totalSequncerConsumeEther: big.NewInt(0),
				endBlockNumber:            big.NewInt(0),
				l2BlockTime:               cfg.l2BlockTime,
			}
		},
	)
}

func (r *reports) Report() {
	fmt.Printf("tps: %d\n", r.tps)
	fmt.Printf("totalConfirmTransactions: %d\n", r.totalConfirmTransactions)
	fmt.Printf("l1GasUsed: %d\n", r.l1GasUsed)
	fmt.Printf("l2GasUsed: %d\n", r.l2GasUsed)
	fmt.Printf("l1GasFee: %d (%f ETH)\n", r.l1GasFee, weiToEther(r.l1GasFee))
	fmt.Printf("l2GasFee: %d (%f ETH)\n", r.l2GasFee, weiToEther(r.l2GasFee))
	fmt.Printf("batcherConsumeEther: %d\n", r.batcherConsumeEther)
	fmt.Printf("proposerConsumeEther: %d\n", r.proposerConsumeEther)
	fmt.Printf("totalSequncerConsumeEther: %d\n", r.totalSequncerConsumeEther)
	fmt.Printf("startBlockNumber: %d\n", r.startBlockNumber)
	fmt.Printf("endBlockNumber: %d\n", r.endBlockNumber)
}
