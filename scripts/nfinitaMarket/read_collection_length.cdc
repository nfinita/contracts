import NfinitaMarket from "../../contracts/NfinitaMarket.cdc"

// This script returns the size of an account's SaleOffer collection.

pub fun main(account: Address): Int {
    let acct = getAccount(account)
    let marketCollectionRef = acct
        .getCapability<&NfinitaMarket.Collection{NfinitaMarket.CollectionPublic}>(
             NfinitaMarket.CollectionPublicPath
        )
        .borrow()
        ?? panic("Could not borrow market collection from market address")

    return marketCollectionRef.getSaleOfferIDs().length
}
