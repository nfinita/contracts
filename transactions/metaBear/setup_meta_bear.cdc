import MetaBear from "../../contracts/MetaBear.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction {
    let collectionData: &MetaBear.CollectionData

    prepare(signer: AuthAccount) {
        self.collectionData = signer.borrow<&MetaBear.CollectionData>(from: MetaBear.CollectionDataPath) ?? panic("Could not borrow Collection Data")
    }

    execute {
        self.collectionData.setCollectionMetadata(metadata: METADATA_PLACEHOLDER)
    }
}
