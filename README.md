## WB Network

To ensure compatibility with existing technology and to leverage the
benefits of a popular community, WB Network has chosen to remain
compatible with all existing smart contracts on Ethereum and Ethereum
tooling. This has been achieved by developing based on a go-ethereum
fork, as the team holds high respect for the excellent work of Ethereum:

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://pkg.go.dev/github.com/ethereum/go-ethereum?tab=doc)

WB Network has rolled out a new EVM-compatible network utilizing 
Proof of Authority (PoA) consensus to enable shorter block times 
and reduced fees. As its development is based on go-ethereum fork, 
you may observe that many toolings, binaries, and documentation are 
based on Ethereum.

### PoA consensus

Proof-of-Authority (PoA) consensus is an efficient consensus algorithm
that was coined by Gavin Wood, a co-founder of the Ethereum
blockchain in 2017. In a PoA consensus, all nodes are
pre-authenticated, which allows using consensus types that provide a
high transaction rate in addition to other benefits.

#### Advantages of PoA consensus

Compared to other consensus types that require proof of spent
computational resources (Proof-of-Work) or an existing "share"
(Proof-of-Stake), PoA consensus has several notable advantages:

* It does not require high-performance hardware like PoW
consensus, which demands nodes to spend computational
resources for solving complex mathematical tasks
* The interval of time at which new blocks are generated is
predictable. For PoW and PoS consensuses, this time varies
* Blocks are generated in a sequence at appointed time intervals by
authorized network nodes, leading to a higher transaction rate
* PoA consensus is tolerant to compromised and malicious nodes,
as long as 51% of nodes are not compromised

#### How PoA consensus works in WB Network

In WB Network, only selected nodes known as validators (validating nodes) can
generate new blocks. These nodes are responsible for maintaining the
blockchain network and the distributed ledger. The blockchain registry
maintains the list of validators, and the order of nodes in this list
determines the sequence in which nodes generate new blocks.

### Quickstart with Docker

#### Prerequisites
Ensure you have Docker installed (get it from https://docs.docker.com/get-docker/) and a compatible operating system (Windows, macOS, or Linux).

#### Start a full node
Open a terminal or command prompt, then enter:

```shell
docker run --rm -it \
  --name wbt -v /Users/alice/wbt:/root \
  -p 8545:8545 -p 30303:30303 \
  whitebit/wbt:0.3.0 --wbt-testnet
```

This will start `geth` in snap-sync mode with a DB memory allowance of 1GB just as the
above command does.  It will also create a persistent volume in your home directory for
saving your blockchain as well as map the default ports.

Do not forget `--http.addr 0.0.0.0`, if you want to access RPC from other containers
and/or hosts. By default, `geth` binds to the local interface and RPC endpoints are not
accessible from the outside.

#### Monitor logs 
You can check if your node has started syncing by looking for the following log messages:
```text
INFO [04-12|09:59:58.554] Imported new block receipts              count=21  elapsed=10.245ms  number=21  hash=f4b1f9..0b33a4 age=1w4d20h size=7.25KiB
INFO [04-12|09:59:58.645] Imported new block headers               count=192 elapsed=105.222ms number=768 hash=0fc521..d533e8 age=1w4d19h
INFO [04-12|09:59:58.670] Imported new block receipts              count=21  elapsed=5.967ms   number=42  hash=f6c9ed..38c629 age=1w4d20h size=7.16KiB
INFO [04-12|09:59:58.673] Imported new block receipts              count=11  elapsed=2.979ms   number=53  hash=c78894..0acd7c age=1w4d20h size=3.84KiB
```

## Building the source

For prerequisites and detailed build instructions please read the geth [Installation Instructions](https://geth.ethereum.org/docs/install-and-build/installing-geth).

Building `geth` requires both a Go (version 1.18 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run

```shell
make geth
```

or, to build the full suite of utilities:

```shell
make all
```

## Executables

The WB Network project comes with several wrappers/executables found in the `cmd`
directory.

|   Command   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
|:-----------:| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`geth`**  | Our main Ethereum CLI client. It is the entry point into the Ethereum network (main-, test- or private net), capable of running as a full node (default), archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Ethereum network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `geth --help` and the [CLI page](https://geth.ethereum.org/docs/interface/command-line-options) for command line options.          |
| `clef`      | Stand-alone signing tool, which can be used as a backend signer for `geth`.  |
|  `devp2p`   | Utilities to interact with nodes on the networking layer, without running a full blockchain. |
|  `abigen`   | Source code generator to convert Ethereum contract definitions into easy to use, compile-time type-safe Go packages. It operates on plain [Ethereum contract ABIs](https://docs.soliditylang.org/en/develop/abi-spec.html) with expanded functionality if the contract bytecode is also available. However, it also accepts Solidity source files, making development much more streamlined. Please see our [Native DApps](https://geth.ethereum.org/docs/dapp/native-bindings) page for details. |
| `bootnode`  | Stripped down version of our Ethereum client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks.                                                                                                                                                                                                                                                                 |
|    `evm`    | Developer utility version of the EVM (Ethereum Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of EVM opcodes (e.g. `evm --code 60ff60ff --debug run`).                                                                                                                                                                                                                                                                     |
|  `rlpdump`  | Developer utility tool to convert binary RLP ([Recursive Length Prefix](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp)) dumps (data encoding used by the Ethereum protocol both network as well as consensus wise) to user-friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`).                                                                                                                                                                                                                                 |
|  `puppeth`  | a CLI wizard that aids in creating a new Ethereum network.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |

## Running `geth`

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://geth.ethereum.org/docs/interface/command-line-options)),
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `geth` instance.

### Hardware Requirements
Minimum:

CPU with 2+ cores
8GB RAM
100GB free storage space to sync the Testnet
8 MBit/sec download Internet service

Recommended (Testnet):
Fast CPU with 4+ cores
16GB+ RAM
High-performance SSD with at least 500GB of free space
25+ MBit/sec download Internet service

### Configuration

As an alternative to passing the numerous flags to the `geth` binary, you can also pass a
configuration file via:

```shell
$ geth --wbt-testnet --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to
export your existing configuration:

```shell
$ geth --wbt-testnet  --your-favourite-flags dumpconfig
```

### Programmatically interfacing `geth` nodes

As a developer, sooner rather than later you'll want to start interacting with `geth` and the
Ethereum network via your own programs and not manually through the console. To aid
this, `geth` has built-in support for a JSON-RPC based APIs ([standard APIs](https://ethereum.github.io/execution-apis/api-documentation/)
and [`geth` specific APIs](https://geth.ethereum.org/docs/rpc/server)).
These can be exposed via HTTP, WebSockets and IPC (UNIX sockets on UNIX based
platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by `geth`,
whereas the HTTP and WS interfaces need to manually be enabled and only expose a
subset of APIs due to security reasons. These can be turned on/off and configured as
you'd expect.

HTTP based JSON-RPC API options:

  * `--http` Enable the HTTP-RPC server
  * `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
  * `--http.port` HTTP-RPC server listening port (default: `8545`)
  * `--http.api` API's offered over the HTTP-RPC interface (default: `eth,net,web3`)
  * `--http.corsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--ws.addr` WS-RPC server listening interface (default: `localhost`)
  * `--ws.port` WS-RPC server listening port (default: `8546`)
  * `--ws.api` API's offered over the WS-RPC interface (default: `eth,net,web3`)
  * `--ws.origins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: `admin,debug,eth,miner,net,personal,txpool,web3`)
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to
connect via HTTP, WS or IPC to a `geth` node configured with the above flags and you'll
need to speak [JSON-RPC](https://www.jsonrpc.org/specification) on all transports. You
can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based
transport before doing so! Hackers on the internet are actively trying to subvert
Ethereum nodes with exposed APIs! Further, all browser tabs can access locally
running web servers, so malicious web pages could try to subvert locally available
APIs!**

## Contribution

Thank you for considering to help out with the source code! We welcome contributions
from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to WB Network, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit
more complex changes though, please write us a Github issue!

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://geth.ethereum.org/docs/developers/geth-developer/dev-guide)
for more details on configuring your environment, managing project dependencies, and
testing procedures.

## License

The wbt library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The wbt binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
