import MetaBear from "../../contracts/MetaBear.cdc"

transaction (
    community: Address,
    communityFee2ndPercentage: UFix64,
    creator: Address,
    creatorFeeMintPercentage: UFix64,
    creatorFee2ndPercentage: UFix64,
    platform: Address,
    platformFeeMintPercentage: UFix64,
    platformFee2ndPercentage: UFix64,
) {
    let collectionData: &MetaBear.CollectionData
    let settings: {String: AnyStruct}

    prepare(signer: AuthAccount) {
        self.collectionData = signer.borrow<&MetaBear.CollectionData>(from: MetaBear.CollectionDataPath)
            ?? panic("Could not borrow Collection Data")
        // CHANGE VALUES ACCORDINGLY.
        self.settings = {
            "community": community,
            "communityFee2ndPercentage": communityFee2ndPercentage,
            "creator": creator,
            "creatorFeeMintPercentage": creatorFeeMintPercentage,
            "creatorFee2ndPercentage": creatorFee2ndPercentage,
            "platform": platform,
            "platformFeeMintPercentage": platformFeeMintPercentage,
            "platformFee2ndPercentage": platformFee2ndPercentage
        }
    }

    execute {
        self.collectionData.setCollectionSettings(settings: self.settings)
    }
}
