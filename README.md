# Metagood

👋 Welcome! This has the information about the project and how you can run our Metagood NFT Marketplace on the Flow blockchain.

- Metagood is an **NFT marketplace** built with [Cadence](https://docs.onflow.org/cadence), Flow's resource-oriented smart contract programming language.
- Currently, each developer is using their own Testnet account to develop on.
- This project is built on top of [kitty-items](https://github.com/onflow/kitty-items) project by Flow. It is a very good boilerplate that speeds up development and understanding of the system.

## ✨ Getting Started

### 1. Install the Flow CLI

Before you start, install the [Flow command-line interface (CLI)](https://docs.onflow.org/flow-cli).

_⚠️ This project requires `flow-cli v0.15.0` or above._

### 2. Clone the project

```sh
git clone https://github.com/nfinita/first-market.git
```

### 3. Create a Flow Testnet account

You'll need a Testnet account to work on this project. Here's how to make one:

#### Generate a key pair 

Generate a new key pair with the Flow CLI:

```sh
flow keys generate
```

_⚠️ Make sure to save these keys in a safe place, you'll need them later._

#### Create your account

Go to the [Flow Testnet Faucet](https://testnet-faucet-v2.onflow.org/) to create a new account. Use the **public key** from the previous step.

#### Save your keys

After your account has been created, save the address and private key to the following environment variables:

```sh
# Replace these values with your own!
export FLOW_ADDRESS=0xabcdef12345689
export FLOW_PRIVATE_KEY=xxxxxxxxxxxx
```

### 4. Deploy the contracts

```sh
flow project deploy --network=testnet
```

If you'd like to look at the contracts in your account, to confirm that everything was deploy properly, you can use the following cli command:
```sh
flow accounts get $FLOW_ADDRESS --network=testnet
```

### 5. Run the API

After the contracts are deployed, follow the [Metagood API instructions](https://github.com/nfinita/first-market/blob/master/api/README.md)
to install and run the Metagood API. This backend service is responsible for initializing accounts, minting NFTs, and processing events.

### 6. Launch the web app

Lastly, follow the [Metagood Web instructions](https://github.com/nfinita/first-market/blob/master/web/README.md) to launch the Metagood front-end React app.

## Project Overview

![Project Overview](https://github.com/onflow/kitty-items/blob/master/kitty-items-diagram.png)

## 🔎 Legend

Above is a basic diagram of the parts of this project contained in each folder, and how each part interacts with the others.

### 1. Web App (Static website) | [first-market/web](https://github.com/nfinita/first-market/tree/master/web)

A true dapp, client-only web app. This is a complete web application built with React that demonstrates how to build a static website that can be deployed to an environment like IPFS and connects directly to the Flow blockchain using `@onflow/fcl`. No servers required. `@onflow/fcl` handles authentication and authorization of [Flow accounts](https://docs.onflow.org/concepts/accounts-and-keys/), [signing transactions](https://docs.onflow.org/concepts/transaction-signing/), and querying data using using Cadence scripts.

### 2. Web Server! | [first-market/api](https://github.com/nfinita/first-market/tree/master/api)

We love decentralization, but servers are still very useful, and this one's no exception. The code in this project demonstrates how to connect to Flow using [Flow JavaScript SDK](https://github.com/onflow/flow-js-sdk) from a Node JS backend. It's also chalk-full of handy patterns you'll probably want to use for more complex and feature-rich blockchain applications, like storing and querying events using a SQL database (Postgres).

### 3. Cadence Code | [first-market/cadence](https://github.com/nfinita/first-market/master/cadence)

[Cadence](https://docs.onflow.org/cadence) smart contracts, scripts & transactions for your viewing pleasure. This folder contains all of the blockchain logic for the marketplace application. Here you will find examples of [fungible token](https://github.com/onflow/flow-ft) and [non-fungible token (NFT)](https://github.com/onflow/flow-nft) smart contract implementations, as well as the scripts and transactions that interact with them. It also contains examples of how to _test_ your Cadence code (tests written in Golang).
