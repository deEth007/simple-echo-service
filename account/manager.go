package account

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

type manager struct {
	*CLIConfig

	file string
}

type distributor struct {
	name            string
	address         common.Address
	client          *ethclient.Client
	pool            []int
	sendTransaction func(common.Address, *big.Int) (common.Hash, error)
}

func (d *distributor) waitTransaction(
	ctx context.Context,
	txHash common.Hash,
	recieptCh chan<- *types.Receipt,
) {
	queryTicker := time.NewTicker(500 * time.Millisecond)
	defer queryTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-queryTicker.C:
			receiept, _ := d.client.TransactionReceipt(ctx, txHash)
			if receiept != nil {
				recieptCh <- receiept
				return
			}
		}
	}
}

func newDistributor(name string, url string, privKey *ecdsa.PrivateKey) (distributor, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return distributor{}, err
	}

	publicKey := privKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return distributor{}, err
	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return distributor{}, err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return distributor{}, err
	}

	sendTx := func(address common.Address, amount *big.Int) (common.Hash, error) {
		var data []byte
		tx := types.NewTransaction(nonce, address, amount, gasLimit, gasPrice, data)

		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainID), privKey)
		if err != nil {
			return common.Hash{}, err
		}

		err = client.SendTransaction(context.Background(), signedTx)
		fmt.Printf("%s sent Tx to %s\n", name, address.Hex())
		if err != nil {
			return common.Hash{}, err
		}

		nonce++

		hash := signedTx.Hash()

		return hash, nil
	}

	return distributor{
		name:            name,
		address:         fromAddress,
		client:          client,
		sendTransaction: sendTx,
	}, nil
}

func Main() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		cliConfig := ReadCLIConfig(ctx)

		mgr, err := newAccountMgr(cliConfig)
		if err != nil {
			return err
		}

		cmd := ctx.Command.Name

		err = mgr.start(cmd)
		if err != nil {
			return err
		}

		return nil
	}
}

func newAccountMgr(config CLIConfig) (*manager, error) {
	return &manager{
		CLIConfig: &config,
		file:      getAccountFilePath(),
	}, nil
}

func (aMgr *manager) start(cmd string) error {
	switch cmd {
	case "generate":
		if err := aMgr.generateAccounts(); err != nil {
			return err
		}
	case "faucet":
		fmt.Println("Faucet start")
		err := aMgr.faucet()
		if err != nil {
			return err
		}
	}
	return nil
}

func (aMgr *manager) generateAccounts() error {
	var keys []string
	for i := int64(0); i < aMgr.CountAccounts; i++ {
		newPrivKey, err := crypto.GenerateKey()
		if err != nil {
			return err
		}
		privateKeyBytes := crypto.FromECDSA(newPrivKey)
		keys = append(keys, hexutil.Encode(privateKeyBytes))
	}

	return write(aMgr.file, keys)
}

func (aMgr *manager) faucet() error {
	stringPrivateKeys, err := read(aMgr.file)
	if err != nil {
		return err
	}

	var testerKeys []*ecdsa.PrivateKey
	var testerAddresses []common.Address
	for _, k := range stringPrivateKeys {
		privKey, err := stringToPrivateKey(k)
		if err != nil {
			return err
		}
		testerKeys = append(testerKeys, privKey)
		testerAddresses = append(testerAddresses, getAddress(privKey))
	}

	var distributors []distributor
	for i := 0; i < 10; i++ {
		d, err := newDistributor(fmt.Sprintf("distributor-%d", i), aMgr.RpcURL, testerKeys[i])
		if err != nil {
			return err
		}
		pool := makePool(i, len(testerAddresses))
		d.pool = pool
		distributors = append(distributors, d)
	}
	fmt.Println("here")

	masterPrivateKey, err := stringToPrivateKey(aMgr.DistributorPrivateKey)
	if err != nil {
		return err
	}

	masterDistributor, err := newDistributor("distributor-master", aMgr.RpcURL, masterPrivateKey)
	amount := new(big.Int)
	// 500,000
	amount.SetString("500000000000000000000000", 10)

	var wg sync.WaitGroup

	ch := make(chan *types.Receipt)

	for _, d := range distributors {
		txHash, _ := masterDistributor.sendTransaction(d.address, amount)

		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
			defer cancel()

			masterDistributor.waitTransaction(ctx, txHash, ch)
		}()

	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		fmt.Printf("confirmed transaction %s\n", r.TxHash.Hex())
	}

	receiptChan := make(chan *types.Receipt)

	for _, d := range distributors {
		amount := new(big.Int)
		// 1,000
		amount.SetString("1000000000000000000000", 10)
		wg.Add(1)
		go func(dtb distributor) {
			defer wg.Done()
			for _, i := range dtb.pool {
				txHash, _ := dtb.sendTransaction(testerAddresses[i], amount)

				wg.Add(1)
				go func() {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
					defer cancel()

					dtb.waitTransaction(ctx, txHash, receiptChan)
				}()
			}
		}(d)
	}

	go func() {
		wg.Wait()
		close(receiptChan)
	}()

	count := 0
	for r := range receiptChan {
		count++
		fmt.Printf("confirmed transaction %d: %s\n", count, r.TxHash.Hex())
	}

	return nil
}

func makePool(n, l int) []int {
	var pool []int
	k := l / 10
	for m := 1; m <= k; m++ {
		index := n + 10*m
		if index >= l {
			break
		}
		pool = append(pool, index)
	}
	return pool
}
