# Metagood

👋 Welcome! This has the information about the project and how you can run our Metagood NFT Marketplace on the Flow blockchain.

- Metagood is an **NFT marketplace** built with [Cadence](https://docs.onflow.org/cadence), Flow's resource-oriented smart contract programming language.
- Currently, each developer is using their own Testnet account to develop on.
- This project is built on top of [kitty-items](https://github.com/onflow/kitty-items) project by Flow. It is a very good boilerplate that speeds up development and understanding of the system.

## Project Overview

![Project Overview](https://github.com/onflow/kitty-items/blob/master/kitty-items-diagram.png)

## 🔎 Legend

Above is a basic diagram of the parts of this project contained in each folder, and how each part interacts with the others.

### 1. Web App

A true dapp, client-only web app. This is a complete web application built with React that demonstrates how to build a static website that can be deployed to an environment like IPFS and connects directly to the Flow blockchain using `@onflow/fcl`. No servers required. `@onflow/fcl` handles authentication and authorization of [Flow accounts](https://docs.onflow.org/concepts/accounts-and-keys/), [signing transactions](https://docs.onflow.org/concepts/transaction-signing/), and querying data using using Cadence scripts.

### 2. Web Server

We love decentralization, but servers are still very useful, and this one's no exception. The code in this project demonstrates how to connect to Flow using [Flow JavaScript SDK](https://github.com/onflow/flow-js-sdk) from a Node JS backend. It's also chalk-full of handy patterns you'll probably want to use for more complex and feature-rich blockchain applications, like storing and querying events using a SQL database (Postgres).

### 3. Cadence Code 

[Cadence](https://docs.onflow.org/cadence) smart contracts, scripts & transactions for your viewing pleasure. This folder contains all of the blockchain logic for the marketplace application. Here you will find examples of [fungible token](https://github.com/onflow/flow-ft) and [non-fungible token (NFT)](https://github.com/onflow/flow-nft) smart contract implementations, as well as the scripts and transactions that interact with them. It also contains examples of how to _test_ your Cadence code (tests written in Golang).
