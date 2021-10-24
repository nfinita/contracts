import MetagoodMarket from "../../contracts/MetagoodMarket.cdc"

transaction(collection: Address, itemID: UInt64) {
    let marketCollection: &MetagoodMarket.Collection

    prepare(signer: AuthAccount) {
        self.marketCollection = signer.borrow<&MetagoodMarket.Collection>(from: MetagoodMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed MetagoodMarket Collection")
    }

    execute {
        let offer <-self.marketCollection.remove(collection: collection, itemID: itemID)
        destroy offer
    }
}
