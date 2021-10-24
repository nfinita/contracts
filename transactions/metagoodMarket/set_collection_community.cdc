import MetagoodMarket from "../../contracts/MetagoodMarket.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction(
    account: Address,
    community: Address,
) {
      let marketSettings: &MetagoodMarket.MarketSettings

      prepare(signer: AuthAccount) {
        self.marketSettings = signer.borrow<&MetagoodMarket.MarketSettings>(
            from: MetagoodMarket.MarketSettingsStoragePath
        ) ?? panic("Unable to borrow Market Settings")
      }

    execute {
        self.marketSettings.setCollectionCommunity(
            address: account, communityAddress: community
        )
    }
}
