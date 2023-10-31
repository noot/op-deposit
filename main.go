package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

const DEFAULT_ETHEREUM_ENDPOINT = "http://localhost:8545"

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "optimism-portal-address",
				Usage: "address of deployed optimism portal contract",
			},
			&cli.StringFlag{
				Name:  "ethereum-endpoint",
				Usage: "ethereum JSON-RPC endpoint",
				Value: DEFAULT_ETHEREUM_ENDPOINT,
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "address to deposit to on L2; if empty, will send to the address of the private key",
			},
			&cli.StringFlag{
				Name:  "private-key",
				Usage: "private key of the account to send from",
			},
			&cli.Float64Flag{
				Name:  "value",
				Usage: "amount of ETH to deposit",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	fmt.Println("connecting to ethereum endpoint", c.String("ethereum-endpoint"))

	ec, err := ethclient.Dial(c.String("ethereum-endpoint"))
	if err != nil {
		return fmt.Errorf("failed to dial ethereum endpoint: %w", err)
	}

	optimismPortalAddress := common.HexToAddress(c.String("optimism-portal-address"))
	optimismPortal, err := bindings.NewOptimismPortal(optimismPortalAddress, ec)
	if err != nil {
		return fmt.Errorf("failed to create OptimismPortal bindings: %w", err)
	}

	to := common.HexToAddress(c.String("to"))
	valueF := new(big.Float).SetFloat64(c.Float64("value"))
	valueF = new(big.Float).Mul(valueF, big.NewFloat(1e18))
	value, _ := valueF.Int(nil)

	// TODO: add data parameter
	gasLimit := uint64(21000) // minimumGasLimit = len(data) * 16 + 21000

	privKey, err := crypto.HexToECDSA(c.String("private-key"))
	if err != nil {
		return fmt.Errorf("failed to create private key from flag: %w", err)
	}

	senderAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	balance, err := ec.BalanceAt(context.Background(), senderAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to get sender balance: %w", err)
	}
	fmt.Println("sender balance (wei)", balance)
	fmt.Println("deposit value (wei)", value)

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(31337))
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
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
		return fmt.Errorf("failed to submit deposit transaction: %w", err)
	}

	for {
		receipt, err := ec.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			continue
		}
		fmt.Println("receipt", receipt)
		if receipt.Status == 0 {
			return fmt.Errorf("transaction failed")
		}

		fmt.Println("transaction succeeded!")
		break
	}

	return nil
}
