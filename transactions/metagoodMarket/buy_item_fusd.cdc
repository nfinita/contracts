import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MetagoodMarket from "../../contracts/MetagoodMarket.cdc"
import FUSD from "../../contracts/FUSD.cdc"

transaction(collection: Address, itemID: UInt64, donation: UFix64, marketCollectionAddress: Address) {
    let paymentVault: @FungibleToken.Vault
    let donationVault: @FungibleToken.Vault
    let platformFeeVault: @FungibleToken.Vault
    let creatorFeeVault: @FungibleToken.Vault
    let communityFeeVault: @FungibleToken.Vault
    let collection: &{NonFungibleToken.Receiver}
    let marketCollection: &MetagoodMarket.Collection{MetagoodMarket.CollectionPublic}

    prepare(acct: AuthAccount) {
        self.marketCollection = getAccount(marketCollectionAddress)
            .getCapability<&MetagoodMarket.Collection{MetagoodMarket.CollectionPublic}>(MetagoodMarket.CollectionPublicPath)
            .borrow() ?? panic("Could not borrow market collection from market address")

        let saleItem = self.marketCollection.borrowSaleItem(collection: collection, itemID: itemID)!
        let price = saleItem.price
        let platformFee = saleItem.platformFee
        let creatorFee = saleItem.creatorFee
        let communityFee = saleItem.communityFee

        let mainFUSDVault = acct.borrow<&FUSD.Vault>(from: /storage/fusdVault)
            ?? panic("Cannot borrow FUSD vault from acct storage")
        self.paymentVault <- mainFUSDVault.withdraw(amount: price - platformFee - creatorFee - communityFee)
        self.platformFeeVault <- mainFUSDVault.withdraw(amount: platformFee)
        self.creatorFeeVault <- mainFUSDVault.withdraw(amount: creatorFee)
        self.communityFeeVault <- mainFUSDVault.withdraw(amount: communityFee)
        self.donationVault <- mainFUSDVault.withdraw(amount: donation)

        self.collection = acct.borrow<&{NonFungibleToken.Receiver}>(
            from: /storage/${pathName}
        ) ?? panic("Cannot borrow collection receiver from acct")
    }

    execute {
        self.marketCollection.purchase(
            collection: collection,
            itemID: itemID,
            buyerCollection: self.collection,
            buyerPayment: <- self.paymentVault,
            buyerDonation: <- self.donationVault,
            buyerPlatformFee: <- self.platformFeeVault,
            buyerCreatorFee: <- self.creatorFeeVault,
            buyerCommunityFee: <- self.communityFeeVault,
        )
    }
}
