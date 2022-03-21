package executor

import (
	"fmt"
	"math/big"

	common "occ-swap-server/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"occ-swap-server/model"
)

type Executor interface {
	GetBlockAndTxEvents(height int64) (*common.BlockAndEventLogs, error)
	GetChainName() string
}

// ===================  SwapStarted =============
var (
	SwapStartedEventName        = "SwapStarted"
	ETH2BSCSwapStartedEventHash = ethcmm.HexToHash("0x7b2b39fe8cb99baf3c533665217a130daefeee1af6329eca59c5bf06a53999ac")
	BSC2ETHSwapStartedEventHash = ethcmm.HexToHash("0x7b2b39fe8cb99baf3c533665217a130daefeee1af6329eca59c5bf06a53999ac")
)

type ETH2BSCSwapStartedEvent struct {
	ERC20Addr ethcmm.Address
	FromAddr  ethcmm.Address
	Amount    *big.Int
	FeeAmount *big.Int
}

func (ev *ETH2BSCSwapStartedEvent) ToSwapStartTxLog(log *types.Log) *model.SwapStartTxLog {
	pack := &model.SwapStartTxLog{
		TokenAddr:   ev.ERC20Addr.String(),
		FromAddress: ev.FromAddr.String(),
		Amount:      ev.Amount.String(),

		FeeAmount: ev.FeeAmount.String(),
		BlockHash: log.BlockHash.Hex(),
		TxHash:    log.TxHash.String(),
		Height:    int64(log.BlockNumber),
	}
	return pack
}

func ParseETH2BSCSwapStartEvent(abi *abi.ABI, log *types.Log) (*ETH2BSCSwapStartedEvent, error) {
	var ev ETH2BSCSwapStartedEvent

	fmt.Printf("parse swap start event block=%s", log.Data)

	err := abi.Unpack(&ev, SwapStartedEventName, log.Data)
	if err != nil {
		return nil, err
	}

	ev.ERC20Addr = ethcmm.BytesToAddress(log.Topics[1].Bytes())
	ev.FromAddr = ethcmm.BytesToAddress(log.Topics[2].Bytes())

	return &ev, nil
}

type BSC2ETHSwapStartedEvent struct {
	FeeAmount   *big.Int
	toChainId   *big.Int
	fromAddress ethcmm.Address
	amount      *big.Int
}

func (ev *BSC2ETHSwapStartedEvent) ToSwapStartTxLog(log *types.Log) *model.SwapStartTxLog {
	pack := &model.SwapStartTxLog{
		FromAddress: ev.fromAddress.String(),
		Amount:      ev.amount.String(),
		ToChainId:   ev.toChainId.String(),

		FeeAmount: ev.FeeAmount.String(),
		BlockHash: log.BlockHash.Hex(),
		TxHash:    log.TxHash.String(),
		Height:    int64(log.BlockNumber),
	}
	return pack
}

func ParseBSC2ETHSwapStartEvent(abi *abi.ABI, log *types.Log) (*BSC2ETHSwapStartedEvent, error) {
	var ev BSC2ETHSwapStartedEvent

	ev.toChainId = log.Topics[1].Big()
	ev.fromAddress = ethcmm.BytesToAddress(log.Topics[2].Bytes())
	ev.amount = log.Topics[3].Big()

	return &ev, nil
}

// =================  SphynxSwapPairRegister ===================
var (
	SwapPairRegisterEventName = "SphynxSwapPairRegister"
	SwapPairRegisterEventHash = ethcmm.HexToHash("0x06101386f3a9dd45570dce2027311173d0e136955e5b912edece89cca5bb526d")
)

type SwapPairRegisterEvent struct {
	Sponsor           ethcmm.Address
	ContractAddr      ethcmm.Address
	BEP20ContractAddr ethcmm.Address
	Name              string
	Symbol            string
	Decimals          uint8
}

func (ev *SwapPairRegisterEvent) ToSwapPairRegisterLog(log *types.Log) *model.SwapPairRegisterTxLog {
	pack := &model.SwapPairRegisterTxLog{
		ERC20Addr: ev.ContractAddr.String(),
		BEP20Addr: ev.BEP20ContractAddr.String(),
		Sponsor:   ev.Sponsor.String(),
		Symbol:    ev.Symbol,
		Name:      ev.Name,
		Decimals:  int(ev.Decimals),

		BlockHash: log.BlockHash.Hex(),
		TxHash:    log.TxHash.String(),
		Height:    int64(log.BlockNumber),
	}
	return pack
}

func ParseSwapPairRegisterEvent(abi *abi.ABI, log *types.Log) (*SwapPairRegisterEvent, error) {
	var ev SwapPairRegisterEvent

	err := abi.Unpack(&ev, SwapPairRegisterEventName, log.Data)
	if err != nil {
		return nil, err
	}
	ev.Sponsor = ethcmm.BytesToAddress(log.Topics[1].Bytes())
	ev.ContractAddr = ethcmm.BytesToAddress(log.Topics[2].Bytes())
	ev.BEP20ContractAddr = ethcmm.BytesToAddress(log.Topics[3].Bytes())
	return &ev, nil
}
