import MetaBear from "../../contracts/MetaBear.cdc"

transaction {
    let collectionData: &MetaBear.CollectionData
    let settings: {String: AnyStruct}

    prepare(signer: AuthAccount) {
        self.collectionData = signer.borrow<&MetaBear.CollectionData>(from: MetaBear.CollectionDataPath)
            ?? panic("Could not borrow Collection Data")
        // CHANGE VALUES ACCORDINGLY.
        self.settings = {
            "community": 0xad5215d0b261af0e as Address,
            "communityFee2ndPercentage": 0.03 as UFix64,
            "creator": 0x2f5e4d202bf9e66d as Address,
            "creatorFeeMintPercentage": 0.08 as UFix64,
            "creatorFee2ndPercentage": 0.02 as UFix64,
            "platform": 0xad5215d0b261af0e as Address,
            "platformFeeMintPercentage": 0.05 as UFix64,
            "platformFee2ndPercentage": 0.025 as UFix64
        }
    }

    execute {
        self.collectionData.setCollectionSettings(settings: self.settings)
    }
}
