import NfinitaMarket from "../../contracts/NfinitaMarket.cdc"

transaction(collection: Address, itemID: UInt64) {
    let marketCollection: &NfinitaMarket.Collection

    prepare(signer: AuthAccount) {
        self.marketCollection = signer.borrow<&NfinitaMarket.Collection>(from: NfinitaMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed NfinitaMarket Collection")
    }

    execute {
        let offer <-self.marketCollection.remove(collection: collection, itemID: itemID)
        destroy offer
    }
}
