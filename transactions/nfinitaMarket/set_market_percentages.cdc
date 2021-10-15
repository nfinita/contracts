import NfinitaMarket from "../../contracts/NfinitaMarket.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction(
    collectionPlatformFeeMintBip: UInt64,
    collectionCreatorFeeMintBip: UInt64,
    collectionPlatformFee2ndBip: UInt64,
    collectionCreatorFee2ndBip: UInt64,
    collectionCommunityFee2ndBip: UInt64,
) {
      let marketSettings: &NfinitaMarket.MarketSettings

      prepare(signer: AuthAccount) {
        self.marketSettings = signer.borrow<&NfinitaMarket.MarketSettings>(
            from: NfinitaMarket.MarketSettingsStoragePath
        ) ?? panic("Unable to borrow Market Settings")
      }

    execute {
        self.marketSettings.setCollectionPlatformFeeMintBip(
            PUBLIC_PATH_PLACEHOLDER, value: collectionPlatformFeeMintBip
        )
        self.marketSettings.setCollectionCreatorFeeMintBip(
            PUBLIC_PATH_PLACEHOLDER, value: collectionCreatorFeeMintBip
        )
        self.marketSettings.setCollectionPlatformFee2ndBip(
            PUBLIC_PATH_PLACEHOLDER, value: collectionPlatformFee2ndBip
        )
        self.marketSettings.setCollectionCreatorFee2ndBip(
            PUBLIC_PATH_PLACEHOLDER, value: collectionCreatorFee2ndBip
        )
        self.marketSettings.setCollectionCommunityFee2ndBip(
            PUBLIC_PATH_PLACEHOLDER, value: collectionCommunityFee2ndBip
        )
    }
}
