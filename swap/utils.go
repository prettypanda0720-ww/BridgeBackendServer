package swap

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"occ-swap-server/common"
	"occ-swap-server/model"
	"occ-swap-server/util"
)

func buildSwapPairInstance(pairs []model.SwapPair) (map[ethcom.Address]*SwapPairIns, error) {
	swapPairInstances := make(map[ethcom.Address]*SwapPairIns, len(pairs))

	for _, pair := range pairs {

		lowBound := big.NewInt(0)
		_, ok := lowBound.SetString(pair.LowBound, 10)
		if !ok {
			panic(fmt.Sprintf("invalid lowBound amount: %s", pair.LowBound))
		}
		upperBound := big.NewInt(0)
		_, ok = upperBound.SetString(pair.UpperBound, 10)
		if !ok {
			panic(fmt.Sprintf("invalid upperBound amount: %s", pair.LowBound))
		}

		swapPairInstances[ethcom.HexToAddress(pair.ERC20Addr)] = &SwapPairIns{
			Symbol:     pair.Symbol,
			Name:       pair.Name,
			Decimals:   pair.Decimals,
			LowBound:   lowBound,
			UpperBound: upperBound,
			BEP20Addr:  ethcom.HexToAddress(pair.BEP20Addr),
			ERC20Addr:  ethcom.HexToAddress(pair.ERC20Addr),
		}

		util.Logger.Infof("Load swap pair, symbol %s, bep20 address %s, erc20 address %s", pair.Symbol, pair.BEP20Addr, pair.ERC20Addr)
	}

	return swapPairInstances, nil
}

func GetKeyConfig(cfg *util.Config) (*util.KeyConfig, error) {
	if cfg.KeyManagerConfig.KeyType == common.AWSPrivateKey {
		result, err := util.GetSecret(cfg.KeyManagerConfig.AWSSecretName, cfg.KeyManagerConfig.AWSRegion)
		if err != nil {
			return nil, err
		}

		keyConfig := util.KeyConfig{}
		err = json.Unmarshal([]byte(result), &keyConfig)
		if err != nil {
			return nil, err
		}
		return &keyConfig, nil
	} else {
		return &util.KeyConfig{
			HMACKey:         cfg.KeyManagerConfig.LocalHMACKey,
			AdminApiKey:     cfg.KeyManagerConfig.LocalAdminApiKey,
			AdminSecretKey:  cfg.KeyManagerConfig.LocalAdminSecretKey,
			BSCPrivateKey:   cfg.KeyManagerConfig.LocalBSCTxHash,
			ETHPrivateKey:   cfg.KeyManagerConfig.LocalETHPrivateKey,
			MATICPrivateKey: cfg.KeyManagerConfig.LocalMATICPrivateKey,
		}, nil
	}
}

func abiEncodeFillETH2BSCSwap(ethTxHash ethcom.Hash, erc20Addr ethcom.Address, toAddress ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("fillETH2BSCSwap", ethTxHash, erc20Addr, toAddress, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeERC20Transfer(recipient ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("transfer", recipient, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeFillBSC2ETHSwap(ethTxHash ethcom.Hash, erc20Addr ethcom.Address, toAddress ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("fillBSC2ETHSwap", ethTxHash, erc20Addr, toAddress, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeFillSwap(toChainId *big.Int, toAddress ethcom.Address, amount *big.Int, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("fillSwap", big.NewInt(0), toChainId, toAddress, amount)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func abiEncodeCreateSwapPair(registerTxHash ethcom.Hash, erc20Addr ethcom.Address, bep20Addr ethcom.Address, name, symbol string, decimals uint8, abi *abi.ABI) ([]byte, error) {
	data, err := abi.Pack("createSwapPair", registerTxHash, erc20Addr, bep20Addr, name, symbol, decimals)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func buildSignedTransaction(contract ethcom.Address, ethClient *ethclient.Client, txInput []byte, privateKey *ecdsa.PrivateKey, chainId *big.Int) (*types.Transaction, error) {
	txOpts := bind.NewKeyedTransactor(privateKey)

	nonce, err := ethClient.PendingNonceAt(context.Background(), txOpts.From)
	if err != nil {
		return nil, err
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	value := big.NewInt(0)
	msg := ethereum.CallMsg{From: txOpts.From, To: &contract, GasPrice: gasPrice, Value: value, Data: txInput}
	gasLimit, err := ethClient.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	fmt.Printf("estimateGas: %s", txInput)

	rawTx := types.NewTransaction(nonce, contract, value, gasLimit, gasPrice, txInput)
	signedTx, err := txOpts.Signer(types.NewEIP155Signer(chainId), txOpts.From, rawTx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("signer: %s", txInput)

	return signedTx, nil
}

func buildNativeCoinTransferTx(contract ethcom.Address, ethClient *ethclient.Client, value *big.Int, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	txOpts := bind.NewKeyedTransactor(privateKey)

	nonce, err := ethClient.PendingNonceAt(context.Background(), txOpts.From)
	if err != nil {
		return nil, err
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	fmt.Printf("gasPrice: %d", gasPrice)
	if err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{From: txOpts.From, To: &contract, GasPrice: gasPrice, Value: value}
	gasLimit, err := ethClient.EstimateGas(context.Background(), msg)
	fmt.Printf("gasLimit: %d", gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	rawTx := types.NewTransaction(nonce, contract, value, gasLimit, gasPrice, nil)
	signedTx, err := txOpts.Signer(types.HomesteadSigner{}, txOpts.From, rawTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func BuildKeys(privateKeyStr string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = privateKeyStr[2:]
	}
	priKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, nil, err
	}
	publicKey, ok := priKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("get public key error")
	}
	return priKey, publicKey, nil
}
