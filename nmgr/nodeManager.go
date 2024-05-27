package nmgr

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-chain-ops/state"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var testBalance = hexutil.MustDecodeBig("0x200000000000000000000000000000000000000000000000000000000000000")

type TypeGenesis string

const (
	l1GenesisName    TypeGenesis = "l1-genesis"
	l2GenesisName    TypeGenesis = "l2-genesis"
	rollupName       TypeGenesis = "rollup"
	addressesNaame   TypeGenesis = "addresses"
	jwtName          TypeGenesis = "jwt"
	deployConfigName TypeGenesis = "deploy-config"
)

type NodeManager interface {
	Start() error
	Stop()
	Destroy() error
}

type BaseNodeManager struct {
	genesisDir string
	env        []string

	*Config
}

func (b *BaseNodeManager) Start() error {
	if err := b.generateJWT(); err != nil {
		return err
	}

	if err := b.updateTimestamp(); err != nil {
		return err
	}

	if err := b.generateL1Genesis(); err != nil {
		return err
	}

	addresses, err := jsonutil.LoadJSON[map[string]interface{}](b.AddressFilePath)
	if err != nil {
		return err
	}

	dir := b.DockerComposeDirPath
	env := []string{
		fmt.Sprintf("L1_GENESIS_FILE_PATH=%s", b.getGenesisFilePath(l1GenesisName)),
		fmt.Sprintf("L2_GENESIS_FILE_PATH=%s", b.getGenesisFilePath(l2GenesisName)),
		fmt.Sprintf("ROLLUP_FILE_PATH=%s", b.getGenesisFilePath(rollupName)),
		fmt.Sprintf("JWT_SECRET_FILE_PATH=%s", b.getGenesisFilePath(jwtName)),
		fmt.Sprintf("L2OO_ADDRESS=%s", (*addresses)["L2OutputOracleProxy"]),
	}
	b.env = env

	if err := runCommand(
		dir, env, "docker", "compose", "up", "-d", "l1"); err != nil {
		return err
	}
	if err := waitUpServer("8545", time.Duration(10*time.Second)); err != nil {
		return err
	}
	if err := waitRPCServer("8545", time.Duration(10*time.Second)); err != nil {
		return err
	}

	if err := b.generateL2Genesis(); err != nil {
		return err
	}

	if err := runCommand(
		dir, env, "docker", "compose", "up", "-d", "l2"); err != nil {
		return err
	}
	if err := waitUpServer("9545", time.Duration(10*time.Second)); err != nil {
		return err
	}
	if err := waitRPCServer("9545", time.Duration(10*time.Second)); err != nil {
		return err
	}

	if err := runCommand(
		dir, env, "docker", "compose", "up", "-d", "op-node", "op-proposer", "op-batcher"); err != nil {
		return err
	}

	return nil
}

func (b *BaseNodeManager) Stop() {}

func (b *BaseNodeManager) Destroy() error {
	dir := b.DockerComposeDirPath

	if err := runCommand(
		dir, b.env, "docker", "compose", "rm"); err != nil {
		return err
	}

	if err := runCommand(
		dir, b.env, "docker", "volume", "prune", "--all"); err != nil {
		return err
	}

	return nil
}

func NewBaseNodeManager(cfg *Config) (*BaseNodeManager, error) {
	gDir, err := makeGenesisDir()
	if err != nil {
		return nil, err
	}

	return &BaseNodeManager{
		genesisDir: gDir,
		Config:     cfg,
	}, nil
}

func (b *BaseNodeManager) generateL1Genesis() error {
	configPath := b.getGenesisFilePath(deployConfigName)
	deployConfig, err := genesis.NewDeployConfig(configPath)
	if err != nil {
		return err
	}

	deployments, err := genesis.NewL1Deployments(b.AddressFilePath)
	if err != nil {
		return err
	}

	if deployments != nil {
		deployConfig.SetDeployments(deployments)
	}

	if err := deployConfig.Check(); err != nil {
		return err
	}

	dump, err := genesis.NewStateDump(b.AllocFilePath)
	if err != nil {
		return err
	}

	l1Genesis, err := genesis.BuildL1DeveloperGenesis(deployConfig, dump, deployments)
	if err != nil {
		return err
	}

	if len(b.FaucetAccounts) > 0 {
		l1Genesis = b.faucet(l1Genesis)
	}

	outputFile := b.getGenesisFilePath(l1GenesisName)

	return jsonutil.WriteJSON(outputFile, l1Genesis, 0o666)
}

func (b *BaseNodeManager) generateL2Genesis() error {
	configPath := b.getGenesisFilePath(deployConfigName)
	deployConfig, err := genesis.NewDeployConfig(configPath)
	if err != nil {
		return err
	}

	deployments, err := genesis.NewL1Deployments(b.AddressFilePath)
	if err != nil {
		return err
	}
	deployConfig.SetDeployments(deployments)

	var l1StartBlock *types.Block
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return err
	}

	if deployConfig.L1StartingBlockTag == nil {
		l1StartBlock, err = client.BlockByNumber(context.Background(), nil)
		if err != nil {
			return err
		}
		tag := rpc.BlockNumberOrHashWithHash(l1StartBlock.Hash(), true)
		deployConfig.L1StartingBlockTag = (*genesis.MarshalableRPCBlockNumberOrHash)(&tag)
	} else if deployConfig.L1StartingBlockTag.BlockHash != nil {
		l1StartBlock, err = client.BlockByHash(context.Background(), *deployConfig.L1StartingBlockTag.BlockHash)
		if err != nil {
			return err
		}
	} else if deployConfig.L1StartingBlockTag.BlockNumber != nil {
		l1StartBlock, err = client.BlockByNumber(context.Background(), big.NewInt(deployConfig.L1StartingBlockTag.BlockNumber.Int64()))
		if err != nil {
			return err
		}
	}

	if l1StartBlock == nil {
		return errors.New("no starting L1 block")
	}

	if err := deployConfig.Check(); err != nil {
		return err
	}

	l2Genesis, err := genesis.BuildL2Genesis(deployConfig, l1StartBlock)
	if err != nil {
		return err
	}

	if len(b.FaucetAccounts) > 0 {
		l2Genesis = b.faucet(l2Genesis)
	}

	l2GenesisBlock := l2Genesis.ToBlock()
	rollupConfig, err := deployConfig.RollupConfig(l1StartBlock, l2GenesisBlock.Hash(), l2GenesisBlock.Number().Uint64())
	if err != nil {
		return err
	}
	if err := rollupConfig.Check(); err != nil {
		return err
	}

	outputL2Genesis := b.getGenesisFilePath(l2GenesisName)
	if err := jsonutil.WriteJSON(outputL2Genesis, l2Genesis, 0o666); err != nil {
		return err
	}

	outputRollup := b.getGenesisFilePath(rollupName)
	return jsonutil.WriteJSON(outputRollup, rollupConfig, 0o666)
}

func (b *BaseNodeManager) generateJWT() error {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return err
	}

	hexString := hex.EncodeToString(randomBytes)

	outputFile := b.getGenesisFilePath(jwtName)
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(hexString))
	if err != nil {
		return err
	}
	return nil
}

func (b *BaseNodeManager) faucet(genesis *core.Genesis) *core.Genesis {
	db := state.NewMemoryStateDB(genesis)
	for _, account := range b.FaucetAccounts {
		if !db.Exist(account) {
			db.CreateAccount(account)
		}
		db.AddBalance(account, testBalance)
	}
	return db.Genesis()
}

func (b *BaseNodeManager) getGenesisFilePath(name TypeGenesis) string {
	return fmt.Sprintf("%s/%s", b.genesisDir, name)
}

func makeGenesisDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	gDir := fmt.Sprintf("%s/%s", h, ".tokamak-trunks")
	if _, err := os.Stat(gDir); os.IsNotExist(err) {
		err := os.Mkdir(gDir, 0755)
		if err != nil {
			return "", err
		}
	}
	return gDir, nil
}

func delGenesisDir(gDir string) error {
	if _, err := os.Stat(gDir); err == nil {
		err := os.RemoveAll(gDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCommand(dir string, env []string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		return err
	} else {
		fmt.Printf("%s\n", output)
	}
	return nil
}

func waitUpServer(port string, timeout time.Duration) error {
	url := fmt.Sprintf("http://localhost:%s", port)
	ch := make(chan bool)
	go func() {
		for {
			_, err := http.Get(url)
			if err == nil {
				ch <- true
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server did not reply after %v", timeout)
	}
}

func waitRPCServer(port string, timeout time.Duration) error {
	url := fmt.Sprintf("http://localhost:%s", port)
	body := []byte(`{"id":1, "jsonrpc":"2.0", "method": "eth_chainId", "params":[]}`)

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	ch := make(chan bool)
	go func() {
		for {
			res, err := client.Do(r)
			if err == nil && res.StatusCode < 300 {
				res.Body.Close()
				ch <- true
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server did not reply after %v", timeout)
	}
}

func (b *BaseNodeManager) updateTimestamp() error {
	deployConfig, err := jsonutil.LoadJSON[map[string]interface{}](b.DeployConfigFilePath)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	unixTime := currentTime.Unix()
	hexTime := fmt.Sprintf("0x%x", unixTime)
	(*deployConfig)["l1GenesisBlockTimestamp"] = hexTime

	output := b.getGenesisFilePath(deployConfigName)

	return jsonutil.WriteJSON(output, *deployConfig, 0o666)
}
