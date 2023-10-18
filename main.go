package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const ETHEREUM_ENDPOINT = "http://localhost:8545"

func main() {
	ec, err := ethclient.Dial(ETHEREUM_ENDPOINT)
	if err != nil {
		panic(err)
	}

	optimismPortalAddress := common.HexToAddress("0xF87a0abe1b875489CA84ab1E4FE47A2bF52C7C64")
	optimismPortal, err := bindings.NewOptimismPortal(optimismPortalAddress, ec)
	if err != nil {
		panic(err)
	}

	to := common.HexToAddress("0xa0Ee7A142d267C1f36714E4a8F75612F20a79720") // anvil account 9
	value := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(16), nil)
	gasLimit := uint64(21000) // minimumGasLimit = len(data) * 16 + 21000

	privKey, err := crypto.HexToECDSA("2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6")
	if err != nil {
		panic(err)
	}
	senderAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	if senderAddr != to {
		panic("senderAddr != to")
	}

	balance, err := ec.BalanceAt(context.Background(), to, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("sender balance (wei)", balance)
	fmt.Println("deposit value (wei)", value)

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(31337))
	if err != nil {
		panic(err)
	}

	tx, err := optimismPortal.DepositTransaction(&bind.TransactOpts{
		Value:    value,
		Signer:   auth.Signer,
		From:     auth.From,
		GasPrice: big.NewInt(100000000000),
		GasLimit: 60000,
		Context:  context.Background(),
	}, to, value, gasLimit, false, []byte{})
	if err != nil {
		panic(err)
	}

	for {
		receipt, err := ec.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			continue
		}
		fmt.Println("receipt", receipt)
		break
	}
}
