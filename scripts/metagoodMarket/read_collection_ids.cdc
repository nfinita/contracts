import MetagoodMarket from "../../contracts/MetagoodMarket.cdc"

// This script returns an array of all the NFT IDs for sale
// in an account's SaleOffer collection.

pub fun main(address: Address): [String] {
    let marketCollectionRef = getAccount(address)
        .getCapability<&MetagoodMarket.Collection{MetagoodMarket.CollectionPublic}>(
            MetagoodMarket.CollectionPublicPath
        )
        .borrow()
        ?? panic("Could not borrow market collection from market address")

    return marketCollectionRef.getSaleOfferIDs()
}
