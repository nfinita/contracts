import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import FUSD from "../../contracts/FUSD.cdc"
import MetagoodMarket from "../../contracts/MetagoodMarket.cdc"

transaction(itemID: UInt64, price: UFix64, charity: Address?, charityPercentage: UFix64, donationAmount: UFix64, expiryDate: UInt64, collection: Address) {
    let fusdVault: Capability<&{FungibleToken.Receiver}>
    let collectionCap: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>
    let marketCollection: &MetagoodMarket.Collection

    prepare(signer: AuthAccount) {
        // We need a provider capability, but one is not provided by default so we create one if needed.
        let collectionProviderPrivatePath = /private/${pathName}

        self.fusdVault = signer.getCapability<&{FungibleToken.Receiver}>(/public/fusdReceiver)!
        assert(self.fusdVault.borrow() != nil, message: "Missing or mis-typed FUSD receiver")

        // --- Access to Collection Permissions Start ---

        if !signer.getCapability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>(collectionProviderPrivatePath)!.check() {
          signer.link<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>(collectionProviderPrivatePath, target: /storage/${pathName})
        }
        if signer.borrow<&{NonFungibleToken.CollectionPublic}>(from: /storage/${pathName}) == nil {
            panic("Nothing is stored in /storage/${pathName}")
        }
        if !signer.getCapability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>(collectionProviderPrivatePath)!.check() {
            panic("Linking did not work!")
        }
        self.collectionCap = signer.getCapability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>(collectionProviderPrivatePath)!
        assert(self.collectionCap.borrow() != nil, message: "Missing or mis-typed Collection provider")

        // --- Access to Collection Permissions End ---

        // --- Init MetagoodMarket Start ---

        if !signer.getCapability<&MetagoodMarket.Collection{MetagoodMarket.CollectionPublic}>(MetagoodMarket.CollectionPublicPath)!.check() {
          if signer.borrow<&MetagoodMarket.Collection>(from: MetagoodMarket.CollectionStoragePath) == nil {
            signer.save(<-MetagoodMarket.createEmptyCollection(), to: MetagoodMarket.CollectionStoragePath)
          }
          signer.unlink(MetagoodMarket.CollectionPublicPath)
          signer.link<&MetagoodMarket.Collection{MetagoodMarket.CollectionPublic}>(MetagoodMarket.CollectionPublicPath, target: MetagoodMarket.CollectionStoragePath)
        }

        // --- Init MetagoodMarket End ---

        self.marketCollection = signer.borrow<&MetagoodMarket.Collection>(from: MetagoodMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed MetagoodMarket Collection")
    }

    execute {
        let bidVault <- FUSD.createEmptyVault()
        let offer <- MetagoodMarket.createSaleOffer (
            sellerItemProvider: self.collectionCap,
            collectionPublicPath: /public/${pathName},
            collection: collection,
            itemID: itemID,
            typeID: 0,
            sellerPaymentReceiver: self.fusdVault,
            bidVault: <- bidVault,
            price: price,
            charity: charity,
            charityPercentage: charityPercentage,
            donationAmount: donationAmount,
            expiryDate: expiryDate,
        )
        self.marketCollection.insert(offer: <-offer)
    }
}