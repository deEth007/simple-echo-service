package nmgr

import (
	"fmt"
	"io"
	"os"

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

	// L1Genesis *core.Genesis
	// L2Genesis *core.Genesis

	config *Config
}

func (b *BaseNodeManager) Start() error {
	if err := b.copyInfoFiles(); err != nil {
		return err
	}
	// cmd := exec.Command("docker", "compose", "up", "-d", "l1")
	// cmd.Dir = "./nmgr/nodes/optimism/"
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin

	// if err := cmd.Run(); err != nil {
	// 	fmt.Println(err)
	// }
	fmt.Println(b.infoDir)
	fmt.Printf("%v\n", b.config)
	fmt.Printf("%v\n", b.nodeInfo)
	return nil
}
func (b *BaseNodeManager) Stop() {}
func (b *BaseNodeManager) Destroy() {
	delInfoDir(b.infoDir)
}

func (b *BaseNodeManager) Faucet(accounts []common.Address) {
	// b.L1Genesis = faucet(b.L1Genesis, accounts)
	// b.L2Genesis = faucet(b.L2Genesis, accounts)
}

func (b *BaseNodeManager) copyInfoFiles() error {
	dstL1 := fmt.Sprintf("%s/%s.json", b.infoDir, l1GenesisName)
	dstL2 := fmt.Sprintf("%s/%s.json", b.infoDir, l2GenesisName)
	dstRollup := fmt.Sprintf("%s/%s.json", b.infoDir, rollupName)
	dstAddr := fmt.Sprintf("%s/%s.json", b.infoDir, addressesNaame)
	dstJwt := fmt.Sprintf("%s/%s.txt", b.infoDir, jwtName)

	if err := copyFile(b.config.L1GenesisFilePath, dstL1); err != nil {
		return err
	}
	if err := copyFile(b.config.L2GenesisFilePath, dstL2); err != nil {
		return err
	}
	if err := copyFile(b.config.RollupConfigFilePath, dstRollup); err != nil {
		return err
	}
	if err := copyFile(b.config.AddressFilePath, dstAddr); err != nil {
		return err
	}
	if err := copyFile(b.config.JwtFilePath, dstJwt); err != nil {
		return err
	}

	b.nodeInfo = &nodeInfo{
		l1Genesis: dstL1,
		l2Genesis: dstL2,
		rollup:    dstRollup,
		address:   dstAddr,
		jwt:       dstJwt,
	}

	return nil
}

func NewBaseNodeManager(cfg *Config) (*BaseNodeManager, error) {
	iDir, err := makeInfoDir()
	if err != nil {
		return nil, err
	}
	return &BaseNodeManager{
		infoDir: iDir,
		config:  cfg,
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
