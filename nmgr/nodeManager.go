package nmgr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/tokamak-network/tokamak-trunks/utils"

	"github.com/ethereum-optimism/optimism/op-chain-ops/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
)

var testBalance = hexutil.MustDecodeBig("0x200000000000000000000000000000000000000000000000000000000000000")

const (
	l1GenesisName  = "l1-genesis"
	l2GenesisName  = "l2-genesis"
	rollupName     = "rollup"
	addressesNaame = "addresses"
	jwtName        = "jwt"
)

type NodeManager interface {
	Start() error
	Stop()
	Destroy()
	Faucet(accounts []common.Address)
}

type nodeInfo struct {
	l1Genesis string
	l2Genesis string
	rollup    string
	address   string
	jwt       string
}

type BaseNodeManager struct {
	infoDir  string
	nodeInfo *nodeInfo

	L1Genesis *core.Genesis
	L2Genesis *core.Genesis

	DockerComposeFileDirPath string
}

func (b *BaseNodeManager) Start() error {
	var addresses map[string]interface{}
	utils.ReadJson(b.nodeInfo.address, &addresses)

	dir := b.DockerComposeFileDirPath
	env := []string{
		fmt.Sprintf("L1_GENESIS_FILE_PATH=%s", b.nodeInfo.l1Genesis),
		fmt.Sprintf("L2_GENESIS_FILE_PATH=%s", b.nodeInfo.l2Genesis),
		fmt.Sprintf("ROLLUP_FILE_PATH=%s", b.nodeInfo.rollup),
		fmt.Sprintf("JWT_SECRET_FILE_PATH=%s", b.nodeInfo.jwt),
		fmt.Sprintf("L2OO_ADDRESS=%s", addresses["L2OutputOracleProxy"]),
	}

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
func (b *BaseNodeManager) Destroy() {
	// delInfoDir(b.infoDir)
}

func (b *BaseNodeManager) Faucet(accounts []common.Address) {
	b.L1Genesis = faucet(b.L1Genesis, accounts)
	b.L2Genesis = faucet(b.L2Genesis, accounts)

	utils.WriteJson(b.nodeInfo.l1Genesis, b.L1Genesis)
	utils.WriteJson(b.nodeInfo.l2Genesis, b.L2Genesis)
}

func NewBaseNodeManager(cfg *Config) (*BaseNodeManager, error) {
	iDir, err := makeInfoDir()
	if err != nil {
		return nil, err
	}
	info, err := copyInfoFiles(iDir, cfg)
	if err != nil {
		return nil, err
	}
	l1Genesis := &core.Genesis{}
	l2Genesis := &core.Genesis{}
	utils.ReadJson(info.l1Genesis, l1Genesis)
	utils.ReadJson(info.l2Genesis, l2Genesis)

	return &BaseNodeManager{
		infoDir:                  iDir,
		nodeInfo:                 info,
		DockerComposeFileDirPath: cfg.DockerComposeFileDirPath,
		L1Genesis:                l1Genesis,
		L2Genesis:                l2Genesis,
	}, nil
}

func faucet(genesis *core.Genesis, accounts []common.Address) *core.Genesis {
	db := state.NewMemoryStateDB(genesis)
	for _, account := range accounts {
		if !db.Exist(account) {
			db.CreateAccount(account)
		}
		db.AddBalance(account, testBalance)
	}
	return db.Genesis()
}

func makeInfoDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	iDir := fmt.Sprintf("%s/%s", h, ".tokamak-trunks")
	if _, err := os.Stat(iDir); os.IsNotExist(err) {
		err := os.Mkdir(iDir, 0755)
		if err != nil {
			return "", err
		}
	}
	return iDir, nil
}

func delInfoDir(infoDir string) error {
	if _, err := os.Stat(infoDir); err == nil {
		err := os.RemoveAll(infoDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyInfoFiles(dir string, cfg *Config) (*nodeInfo, error) {
	dstL1 := fmt.Sprintf("%s/%s.json", dir, l1GenesisName)
	dstL2 := fmt.Sprintf("%s/%s.json", dir, l2GenesisName)
	dstRollup := fmt.Sprintf("%s/%s.json", dir, rollupName)
	dstAddr := fmt.Sprintf("%s/%s.json", dir, addressesNaame)
	dstJwt := fmt.Sprintf("%s/%s.txt", dir, jwtName)

	if err := copyFile(cfg.L1GenesisFilePath, dstL1); err != nil {
		return nil, err
	}
	if err := copyFile(cfg.L2GenesisFilePath, dstL2); err != nil {
		return nil, err
	}
	if err := copyFile(cfg.RollupConfigFilePath, dstRollup); err != nil {
		return nil, err
	}
	if err := copyFile(cfg.AddressFilePath, dstAddr); err != nil {
		return nil, err
	}
	if err := copyFile(cfg.JwtFilePath, dstJwt); err != nil {
		return nil, err
	}

	return &nodeInfo{
		l1Genesis: dstL1,
		l2Genesis: dstL2,
		rollup:    dstRollup,
		address:   dstAddr,
		jwt:       dstJwt,
	}, nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
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
