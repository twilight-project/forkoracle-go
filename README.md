# btc-oracle
 
This package fetches latest bitcoin chaintip from [forkscanner](https://github.com/twilight-project/forkscanner) and forwards it to the nyks testnet as MsgSeenBtcChaintip message.

This oracle runs common prefix and validity heuristic checks. This also maintains a separate key pair to propose or vote on valid block proposal. For the purposes of validity gadget, we do not need double spend detection or other secondary checks from ForkScanner. 

# Setup

btc-oracle is designed to be used by a nyks validator. In order to use the nyks-cli, a node should be running on the same machine as the btc-oracle.

This package needs `set-delegate-address` command to be run on the nyksd-cli from a validator to map the valdiator's keys (validator and orchestrator) along with the bitcoin public key inside the node database.

Following sample command can be used by replacing the blocks for validator-address (ValAddress), orchestrator-address (AccAddress) and your bitcoin-public-key:

```
nyksd tx nyks set-delegate-addresses [validator-address] [orchestrator-address] [bitcoin-public-key] --from validator-sgp --chain-id nyks --keyring-backend test
```

# Use

To run the btc-oracle, simply setup your local go enviornment and then run `go build`, execute the binary using `./btc-oracle your-validator-name`. Please make sure that forkscanner is running before you start btc-oracle as it tries to connect with the forkscanner via websocket.
