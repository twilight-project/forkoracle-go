# forkoracle-go
 
This package fetches latest bitcoin chaintip from [forkscanner](https://github.com/twilight-project/forkscanner) and passes it on to the nyks testnet as MsgSeenBtcChaintip message.


# Setup

Forkoracle-go is designed to be used by a nyks validator. In order to use the nyks-cli, a node should be running on the same machine as the forkoracle-go.

This package needs `set-delegate-address` command to be run on the nyksd-cli from a validator to map the valdiator's keys (validator and orchestrator) along with the bitcoin public key inside the node database.

Following sample command can be used by replacing the blocks for validator-address (ValAddress), orchestrator-address (AccAddress) and your bitcoin-public-key:

```
nyksd tx nyks set-delegate-addresses [validator-address] [orchestrator-address] [bitcoin-public-key] --from validator-sgp --chain-id nyks --keyring-backend test
```

# Use

To run the forkoracle, simply setup your local go enviornment and then run `go build`, execute the binary using `./forkoracle-go your-validator-name`. Please make sure that forkscanner is running before you start forkoracle-go as it tries to connect with the forkscanner via websocket.
