import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import FUSD from "../../contracts/FUSD.cdc"
import NfinitaMarket from "../../contracts/NfinitaMarket.cdc"

transaction(itemID: UInt64, price: UFix64, charity: Address?, charityPercentage: UFix64, donationAmount: UFix64, expiryDate: UInt64, collection: Address) {
    let fusdVault: Capability<&{FungibleToken.Receiver}>
    let collectionCap: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>
    let marketCollection: &NfinitaMarket.Collection

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

        // --- Init NfinitaMarket Start ---

        if !signer.getCapability<&NfinitaMarket.Collection{NfinitaMarket.CollectionPublic}>(NfinitaMarket.CollectionPublicPath)!.check() {
          if signer.borrow<&NfinitaMarket.Collection>(from: NfinitaMarket.CollectionStoragePath) == nil {
            signer.save(<-NfinitaMarket.createEmptyCollection(), to: NfinitaMarket.CollectionStoragePath)
          }
          signer.unlink(NfinitaMarket.CollectionPublicPath)
          signer.link<&NfinitaMarket.Collection{NfinitaMarket.CollectionPublic}>(NfinitaMarket.CollectionPublicPath, target: NfinitaMarket.CollectionStoragePath)
        }

        // --- Init NfinitaMarket End ---

        self.marketCollection = signer.borrow<&NfinitaMarket.Collection>(from: NfinitaMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed NfinitaMarket Collection")
    }

    execute {
        let bidVault <- FUSD.createEmptyVault()
        let offer <- NfinitaMarket.createSaleOffer (
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