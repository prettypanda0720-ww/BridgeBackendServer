package swap

import (
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"

	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"

	"occ-swap-server/common"
	"occ-swap-server/util"
)

const (
	SwapTokenReceived common.SwapStatus = "received"
	SwapQuoteRejected common.SwapStatus = "rejected"
	SwapConfirmed     common.SwapStatus = "confirmed"
	SwapSending       common.SwapStatus = "sending"
	SwapSent          common.SwapStatus = "sent"
	SwapSendFailed    common.SwapStatus = "sent_fail"
	SwapSuccess       common.SwapStatus = "sent_success"

	SwapPairReceived   common.SwapPairStatus = "received"
	SwapPairConfirmed  common.SwapPairStatus = "confirmed"
	SwapPairSending    common.SwapPairStatus = "sending"
	SwapPairSent       common.SwapPairStatus = "sent"
	SwapPairSendFailed common.SwapPairStatus = "sent_fail"
	SwapPairSuccess    common.SwapPairStatus = "sent_success"
	SwapPairFinalized  common.SwapPairStatus = "finalized"

	RetrySwapConfirmed  common.RetrySwapStatus = "confirmed"
	RetrySwapSending    common.RetrySwapStatus = "sending"
	RetrySwapSent       common.RetrySwapStatus = "sent"
	RetrySwapSendFailed common.RetrySwapStatus = "sent_fail"
	RetrySwapSuccess    common.RetrySwapStatus = "sent_success"

	SwapEth2BSC   common.SwapDirection = "eth_bsc"
	SwapEth2MATIC common.SwapDirection = "eth_matic"
	SwapBSC2Eth   common.SwapDirection = "bsc_eth"
	SwapBSC2MATIC common.SwapDirection = "bsc_matic"
	SwapMATIC2BSC common.SwapDirection = "matic_bsc"
	SwapMATIC2Eth common.SwapDirection = "matic_eth"

	BatchSize                = 50
	TrackSentTxBatchSize     = 100
	SleepTime                = 5
	SwapSleepSecond          = 2
	TrackSwapPairSMBatchSize = 5

	TxFailedStatus = 0x00

	MaxUpperBound = "999999999999999999999999999999999999"
)

var ethClientMutex sync.RWMutex
var bscClientMutex sync.RWMutex
var maticClientMutex sync.RWMutex

type SwapEngine struct {
	mutex    sync.RWMutex
	db       *gorm.DB
	hmacCKey string
	config   *util.Config
	// key is the bsc contract addr
	swapPairsFromERC20Addr map[ethcom.Address]*SwapPairIns
	ethClient              *ethclient.Client
	bscClient              *ethclient.Client
	maticClient            *ethclient.Client
	ethPrivateKey          *ecdsa.PrivateKey
	bscPrivateKey          *ecdsa.PrivateKey
	maticPrivateKey        *ecdsa.PrivateKey
	ethChainID             int64
	bscChainID             int64
	maticChainID           int64
	bep20ToERC20           map[ethcom.Address]ethcom.Address
	erc20ToBEP20           map[ethcom.Address]ethcom.Address

	swapAgentABI *abi.ABI

	ethSwapAgent   ethcom.Address
	bscSwapAgent   ethcom.Address
	maticSwapAgent ethcom.Address
}

type SwapPairEngine struct {
	mutex   sync.RWMutex
	db      *gorm.DB
	hmacKey string
	config  *util.Config

	swapEngine *SwapEngine

	bscClient       *ethclient.Client
	bscPrivateKey   *ecdsa.PrivateKey
	bscChainID      int64
	bscTxSender     ethcom.Address
	bscSwapAgent    ethcom.Address
	bscSwapAgentABi *abi.ABI
}

type SwapPairIns struct {
	Symbol     string
	Name       string
	Decimals   int
	LowBound   *big.Int
	UpperBound *big.Int

	BEP20Addr ethcom.Address
	ERC20Addr ethcom.Address
}
