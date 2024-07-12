package reporter

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var once sync.Once

type reportManager struct {
	w *os.File
}

var reportMgr *reportManager

func (rm *reportManager) Report(r vegeta.Reporter, title string) error {
	rm.w.WriteString(title + "\n")
	err := r.Report(rm.w)
	if err != nil {
		return err
	}
	rm.w.WriteString("\n")
	return nil
}

func (rm *reportManager) Close() {
	reportMgr.w.Close()
}

func GetReportManager() *reportManager {
	return reportMgr
}

type reports struct {
	tps                      *big.Int
	totalConfirmTransactions *big.Int
	l1GasUsed                *big.Int
	l2GasUsed                *big.Int
	blobGasUsed              *big.Int
	blobFee                  *big.Int
	l1Fee                    *big.Int
	l2Fee                    *big.Int
	l1GasPrice               *big.Int
	blobGasPrice             *big.Int
	l2GasPrice               *big.Int
	cumulativeL1GasPrice     *big.Int
	cumulativeBlobGasPrice   *big.Int
	cumulativeL2GasPrice     *big.Int
	startBlockNumber         *big.Int
	endBlockNumber           *big.Int
	l2BlockTime              *big.Int
	receiptCount             uint64
	blobTxCount              uint64
}

var (
	trunksReport *reports
	first        bool
)

func (r *reports) RecordTPS(client *ethclient.Client) {
	startBlock, _ := client.BlockByNumber(context.Background(), r.startBlockNumber)
	endBlock, _ := client.BlockByNumber(context.Background(), r.endBlockNumber)

	d := endBlock.Time() - startBlock.Time()
	duration := new(big.Int).SetUint64(d)

	tr := new(big.Int).Set(r.totalConfirmTransactions)
	if duration.Cmp(big.NewInt(0)) == 0 {
		r.tps = tr.Div(tr, r.l2BlockTime)
		return
	}

	r.tps = tr.Div(tr, duration)
}

func (r *reports) RecordConfirmRequest() {
	r.totalConfirmTransactions.Add(r.totalConfirmTransactions, big.NewInt(1))
}

func (r *reports) RecordReceipt(receipt *types.Receipt) {
	r.receiptCount++
	if receipt.Type == types.BlobTxType {
		r.blobTxCount++
		r.recordBlobGasPrice(receipt)
		r.recordBlobGasUsed(receipt)
		r.recordBlobFee(receipt)
	}
	r.recordStartToLastBlock(receipt)
	r.recordL1GasUsed(receipt)
	r.recordL1Fee(receipt)
	r.recordL2GasUsed(receipt)
	r.recordL2Fee(receipt)
	r.recordL1GasPrice(receipt)
	r.recordL2GasPrice(receipt)
}

func (r *reports) recordStartToLastBlock(receipt *types.Receipt) {
	if first {
		r.startBlockNumber.Set(receipt.BlockNumber)
		first = false
		return
	}
	if r.startBlockNumber.Cmp(receipt.BlockNumber) > 0 {
		r.startBlockNumber.Set(receipt.BlockNumber)
	}
	if r.endBlockNumber.Cmp(receipt.BlockNumber) < 0 {
		r.endBlockNumber.Set(receipt.BlockNumber)
	}
}

func (r *reports) recordBlobGasUsed(receipt *types.Receipt) {
	r.blobGasUsed.Add(r.blobGasUsed, new(big.Int).SetUint64(receipt.BlobGasUsed))
}

func (r *reports) recordBlobFee(receipt *types.Receipt) {
	blobFee := receipt.BlobGasUsed * receipt.BlobGasPrice.Uint64()
	r.blobFee.Add(r.blobFee, new(big.Int).SetUint64(blobFee))
}

func (r *reports) recordL1GasUsed(receipt *types.Receipt) {
	r.l1GasUsed.Add(r.l1GasUsed, receipt.L1GasUsed)
}

func (r *reports) recordL2GasUsed(receipt *types.Receipt) {
	r.l2GasUsed.Add(r.l2GasUsed, new(big.Int).SetUint64(receipt.GasUsed))
}

func (r *reports) recordL1Fee(receipt *types.Receipt) {
	r.l1Fee.Add(r.l1Fee, receipt.L1Fee)
}

func (r *reports) recordL2Fee(receipt *types.Receipt) {
	l2GasFee := big.NewInt(0)
	l2GasFee.Mul(
		receipt.EffectiveGasPrice,
		new(big.Int).SetUint64(receipt.GasUsed),
	)

	r.l2Fee.Add(r.l2Fee, l2GasFee)
}

func (r *reports) recordL1GasPrice(receipt *types.Receipt) {
	r.cumulativeL1GasPrice.Add(r.cumulativeL1GasPrice, receipt.L1GasPrice)
}

func (r *reports) recordBlobGasPrice(receipt *types.Receipt) {
	r.cumulativeBlobGasPrice.Add(r.cumulativeBlobGasPrice, receipt.BlobGasPrice)
}

func (r *reports) recordL2GasPrice(receipt *types.Receipt) {
	r.cumulativeL2GasPrice.Add(r.cumulativeL2GasPrice, receipt.EffectiveGasPrice)
}

func (r *reports) calcGasPrices() {
	if r.receiptCount > 0 {
		r.l1GasPrice.Quo(r.cumulativeL1GasPrice, new(big.Int).SetUint64(r.receiptCount))
		r.l2GasPrice.Quo(r.cumulativeL2GasPrice, new(big.Int).SetUint64(r.receiptCount))
	}
	if r.blobTxCount > 0 {
		r.blobGasPrice.Quo(r.cumulativeBlobGasPrice, new(big.Int).SetUint64(r.blobTxCount))
	}
}

func weiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float).SetInt(wei)
	wetiToEtherFactor := new(big.Float).SetInt(big.NewInt(params.Ether))
	ether.Quo(ether, wetiToEtherFactor)
	return ether
}

func weiToGwei(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.GWei))
}

func GetTrunksReport() *reports {
	return trunksReport
}

func InitReporter(cfg *Config) {
	once.Do(
		func() {
			trunksReport = &reports{
				tps:                      big.NewInt(0),
				totalConfirmTransactions: big.NewInt(0),
				l1GasUsed:                big.NewInt(0),
				l2GasUsed:                big.NewInt(0),
				blobGasUsed:              big.NewInt(0),
				blobFee:                  big.NewInt(0),
				l1Fee:                    big.NewInt(0),
				l2Fee:                    big.NewInt(0),
				l1GasPrice:               big.NewInt(0),
				blobGasPrice:             big.NewInt(0),
				l2GasPrice:               big.NewInt(0),
				cumulativeL1GasPrice:     big.NewInt(0),
				cumulativeBlobGasPrice:   big.NewInt(0),
				cumulativeL2GasPrice:     big.NewInt(0),
				startBlockNumber:         big.NewInt(0),
				endBlockNumber:           big.NewInt(0),
				l2BlockTime:              cfg.l2BlockTime,
			}
			file, _ := os.Create(cfg.filename)
			reportMgr = &reportManager{
				w: file,
			}
			first = true
		},
	)
}

func (r *reports) report(w io.Writer) error {
	r.calcGasPrices()

	const fmtstr = "TPS\t%d\n" +
		"Total Confirmed Tx\t%d\n" +
		"First Confirmed Block Number\t%d\n" +
		"Last Confirmed Block Number\t%d\n" +
		"Total Using Gas\n" +
		"  L1Gas\t%d\n" +
		"  BlobGas\t%d\n" +
		"  L2Gas\t%d\n" +
		"Average Gas Price\n" +
		"  L1Gas\t%d Wei (%f Gwei)\n" +
		"  BlobGas\t%d Wei (%f Gwei)\n" +
		"  L2Gas\t%d Wei (%f Gwei)\n" +
		"Total Fee\n" +
		"  L1Fee\t%d Wei (%f ETH)\n" +
		"  BlobFee\t%d Wei (%f ETH)\n" +
		"  L2Fee\t%d Wei (%f ETH)\n" +
		"L2BlockTime\t%ds\n"
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', tabwriter.StripEscape)
	if _, err := fmt.Fprintf(tw, fmtstr,
		r.tps,
		r.totalConfirmTransactions,
		r.startBlockNumber,
		r.endBlockNumber,
		r.l1GasUsed, r.blobGasUsed, r.l2GasUsed,
		r.l1GasPrice, weiToGwei(r.l1GasPrice), r.blobGasPrice, weiToGwei(r.blobGasPrice), r.l2GasPrice, weiToGwei(r.l2GasPrice),
		r.l1Fee, weiToEther(r.l1Fee), r.blobFee, weiToEther(r.blobFee), r.l2Fee, weiToEther(r.l2Fee),
		r.l2BlockTime,
	); err != nil {
		return err
	}
	return tw.Flush()
}

func TrunksReporter() vegeta.Reporter {
	return func(w io.Writer) (err error) {
		return trunksReport.report(w)
	}
}
