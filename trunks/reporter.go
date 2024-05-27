package trunks

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type Reports struct {
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
}

var once sync.Once

var ChainReporter *Reports

func (r *Reports) RecordTPS(l2BlockTime *big.Int) {
	sb := new(big.Int).Set(r.startBlockNumber)
	eb := new(big.Int).Set(r.endBlockNumber)
	tb := eb.Sub(eb, sb)
	duration := tb.Mul(tb, l2BlockTime)

	tr := new(big.Int).Set(r.totalConfirmTransactions)
	r.tps = tr.Div(tr, duration)
}

func (r *Reports) RecordConfirmRequest() {
	r.totalConfirmTransactions.Add(r.totalConfirmTransactions, big.NewInt(1))
}

func (r *Reports) RecordStartToLastBlock(receipt *types.Receipt) {
	if r.startBlockNumber.Cmp(receipt.BlockNumber) > 0 {
		r.startBlockNumber.Set(receipt.BlockNumber)
	}
	if r.endBlockNumber.Cmp(receipt.BlockNumber) < 0 {
		r.endBlockNumber.Set(receipt.BlockNumber)
	}
}

func (r *Reports) RecordL1GasUsed(receipt *types.Receipt) {
	r.l1GasUsed.Add(r.l1GasUsed, receipt.L1GasUsed)
}

func (r *Reports) RecordL2GasUsed(receipt *types.Receipt) {
	r.l2GasUsed.Add(r.l2GasUsed, new(big.Int).SetUint64(receipt.GasUsed))
}

func (r *Reports) RecordL1GasFee(receipt *types.Receipt) {
	r.l1GasFee.Add(r.l1GasFee, receipt.L1Fee)
}

func (r *Reports) RecordL2GasFee(receipt *types.Receipt) {
	l2GasFee := receipt.EffectiveGasPrice.Mul(receipt.EffectiveGasPrice, new(big.Int).SetUint64(receipt.GasUsed))
	r.l2GasFee.Add(r.l2GasFee, l2GasFee)
}

func weiToEther(val *big.Int) *big.Int {
	return new(big.Int).Div(val, big.NewInt(params.Ether))
}

func init() {
	if ChainReporter == nil {
		once.Do(
			func() {
				ChainReporter = &Reports{
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
				}
			})
	} else {
		fmt.Println("chainReporter already created")
	}
}

func (r *Reports) PrintReport() {
	fmt.Printf("tps: %d\n", r.tps)
	fmt.Printf("totalConfirmTransactions: %d\n", r.totalConfirmTransactions)
	fmt.Printf("l1GasUsed: %d\n", r.l1GasUsed)
	fmt.Printf("l2GasUsed: %d\n", r.l2GasUsed)
	fmt.Printf("l1GasFee: %d\n", r.l1GasFee)
	fmt.Printf("l2GasFee: %d\n", r.l2GasFee)
	fmt.Printf("batcherConsumeEther: %d\n", r.batcherConsumeEther)
	fmt.Printf("proposerConsumeEther: %d\n", r.proposerConsumeEther)
	fmt.Printf("totalSequncerConsumeEther: %d\n", r.totalSequncerConsumeEther)
	fmt.Printf("startBlockNumber: %d\n", r.startBlockNumber)
	fmt.Printf("endBlockNumber: %d\n", r.endBlockNumber)
}
