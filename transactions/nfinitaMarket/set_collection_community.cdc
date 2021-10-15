import NfinitaMarket from "../../contracts/NfinitaMarket.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction(
    account: Address,
    community: Address,
) {
      let marketSettings: &NfinitaMarket.MarketSettings

      prepare(signer: AuthAccount) {
        self.marketSettings = signer.borrow<&NfinitaMarket.MarketSettings>(
            from: NfinitaMarket.MarketSettingsStoragePath
        ) ?? panic("Unable to borrow Market Settings")
      }

    execute {
        self.marketSettings.setCollectionCommunity(
            address: account, communityAddress: community
        )
    }
}
