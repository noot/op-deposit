# op-deposit

simple CLI for making a deposit to the op-stack via `OptimismPortal.sol`.

Usage:
```go
git clone https://github.com/noot/op-deposit && cd op-deposit
go build
./op-deposit --optimism-portal-address=<ADDRESS_OF_DEPLOYED_CONTRACT> --private-key=<PRIVATE-KEY-HEX-STRING> --value=<VALUE-IN-ETH>
``````

Note the value is in ETH, so you can use a float like `0.999` for example.
