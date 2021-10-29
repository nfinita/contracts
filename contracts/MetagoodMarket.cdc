// SPDX-License-Identifier: MIT

import FungibleToken from "./FungibleToken.cdc"
import NonFungibleToken from "./NonFungibleToken.cdc"

/*
    This is a simple NFT initial sale contract for the DApp to use
    in order to list and sell NFTs.

    Its structure is neither what it would be if it was the simplest possible
    market contract or if it was a complete general purpose market contract.
    Rather it's the simplest possible version of a more general purpose
    market contract that indicates how that contract might function in
    broad strokes. This has been done so that integrating with this contract
    is a useful preparatory exercise for code that will integrate with the
    later more general purpose market contract.

    It allows:
    - Anyone to create Sale Offers and place them in a collection, making it
      publicly accessible.
    - Anyone to accept the offer and buy the item.

    It notably does not handle:
    - Multiple different sale NFT contracts.
    - Multiple different payment FT contracts.
    - Splitting sale payments to multiple recipients.

 */

pub contract MetagoodMarket {
    // SaleOffer events.
    //
    // A sale offer has been created.
    pub event SaleOfferCreated(itemID: UInt64, collection: Address, price: UFix64)
    // Someone has purchased an item that was offered for sale.
    pub event SaleOfferAccepted(itemID: UInt64)
    // A sale offer has been destroyed, with or without being accepted.
    pub event SaleOfferFinished(itemID: UInt64)

    // Collection events.
    //
    // A sale offer has been removed from the collection of Address.
    pub event CollectionRemovedSaleOffer(collection: Address, itemID: UInt64, owner: Address)

    // A sale offer has been inserted into the collection of Address.
    pub event CollectionInsertedSaleOffer(
        itemID: UInt64,
        collection: Address,
        typeID: UInt64,
        owner: Address,
        price: UFix64,
    )

    // A sale offer has been purchased.
    pub event SaleOfferPurchased(
        itemID: UInt64,
        collection: Address,
        from_: Address,
        to_: Address,
        price: UFix64,
    )

    // A platform fee for a sale offer has been paid
    pub event SaleOfferPlatformFeePaid(
        itemID: UInt64,
        collection: Address,
        from_: Address,
        platformFee: UFix64,
    )

    // A sale offer has been donated to.
    pub event SaleOfferDonated(
        itemID: UInt64,
        from_: Address,
        to_: Address?,
        amount: UFix64
    )

    // Named paths
    //
    pub let CollectionStoragePath: StoragePath
    pub let CollectionPublicPath: PublicPath

    // Platform FUSD Vault
    pub let platformVaultCapability: Capability<&{FungibleToken.Receiver}>

    // SaleOfferPublicView
    // An interface providing a read-only view of a SaleOffer
    //
    pub resource interface SaleOfferPublicView {
        pub let itemID: UInt64
        pub let collection: Address
        pub let typeID: UInt64
        pub(set) var price: UFix64
        pub let platformFee: UFix64
        pub let creatorFee: UFix64
        pub let communityFee: UFix64
        pub let charity: Address?
        pub let charityPercentage: UFix64
        pub(set) var donationAmount: UFix64
    }

    // SaleOfferBid
    // A bid for an NFT being offered to sale for a set price and donation amount in FUSD.
    //
    pub struct SaleOfferBid {
        pub let price: UFix64;
        pub let donation: UFix64;

        init(price: UFix64, donation: UFix64) {
            self.price = price;
            self.donation = donation;
        }
    }

    // SaleOffer
    // An NFT being offered to sale for a set fee paid in FUSD.
    //
    pub resource SaleOffer: SaleOfferPublicView {
        // Whether the sale has completed with someone purchasing the item.
        pub var saleCompleted: Bool

        // The NFT ID for sale.
        pub let itemID: UInt64

        // The collection which the NFT is a part of
        pub let collection: Address

        // The 'type' of NFT
        pub let typeID: UInt64

        // The sale payment price.
        pub(set) var price: UFix64

        // The sale platform fee amounts
        pub let platformFee: UFix64
        pub let creatorFee: UFix64
        pub let communityFee: UFix64

        // The collection containing that ID.
        access(self) let sellerItemProvider: Capability<&{NonFungibleToken.Provider}>

        // The FUSD vault that will receive that payment if teh sale completes successfully.
        access(self) let sellerPaymentReceiver: Capability<&{FungibleToken.Receiver}>

        // Charity parameters
        pub let charity: Address?
        pub let charityPercentage: UFix64
        pub(set) var donationAmount: UFix64

        // Recipient's Receiver Capabilities
        pub(set) var recipientVaultCap: Capability<&{FungibleToken.Receiver}>?
        pub(set) var recipientCollection: Capability<&NonFungibleToken.Collection{NonFungibleToken.CollectionPublic}>?

        // Platform, Creator and Community Vaults
        pub let platformVaultCapability: Capability<&{FungibleToken.Receiver}>
        pub let creatorVaultCapability: Capability<&{FungibleToken.Receiver}>
        pub let communityVaultCapability: Capability<&{FungibleToken.Receiver}>

        // updateRecipientVaultCap updates the bidder's Vault capability, providing the
        // us with a way to return their FungibleTokens
        access(contract) fun updateRecipient(vaultCap: Capability<&{FungibleToken.Receiver}>, collection: Capability<&NonFungibleToken.Collection{NonFungibleToken.CollectionPublic}>) {
            self.recipientVaultCap = vaultCap
            self.recipientCollection = collection
        }

        // Called by a purchaser to accept the sale offer.
        // If they send the correct payment in FUSD, and if the item is still available,
        // the NFT will be placed in their Collection.
        //
        access(contract) fun accept(
            buyerCollection: &{NonFungibleToken.Receiver},
            buyerPayment: @FungibleToken.Vault,
            buyerDonation: @FungibleToken.Vault,
            buyerPlatformFee: @FungibleToken.Vault,
            buyerCreatorFee: @FungibleToken.Vault,
            buyerCommunityFee: @FungibleToken.Vault,
        ) {
            pre {
                buyerPayment.balance == self.price - self.platformFee - self.creatorFee - self.communityFee:
                    "Buyer's payment does not equal the offer price."
                buyerPlatformFee.balance == self.platformFee:
                    "Buyer's set platform fee does not equal the offer platform fee."
                buyerCreatorFee.balance == self.creatorFee:
                    "Buyer's set creator fee does not equal the offer creator fee."
                buyerCommunityFee.balance == self.communityFee:
                    "Buyer's set community fee does not equal the offer community fee."
                self.saleCompleted == false: "This sale offer has already been accepted."
            }

            self.saleCompleted = true

            let sellerVaultCapability = self.sellerPaymentReceiver.borrow()!
            let communityVault = self.communityVaultCapability.borrow()
                ?? panic("Could not borrow community fee vault")
            let creatorVault = self.creatorVaultCapability.borrow()
                ?? panic("Could not borrow creator fee vault")
            let platformVault = self.platformVaultCapability.borrow()
                ?? panic("Could not borrow platform fee vault")

            sellerVaultCapability.deposit(
                from: <-buyerPayment
            )
            platformVault.deposit(
                from: <-buyerPlatformFee
            )
            creatorVault.deposit(
                from: <-buyerCreatorFee
            )
            communityVault.deposit(
                from: <-buyerCommunityFee
            )

            if self.charity != nil {
                // Send tokens to charity
                let charityVault = getAccount(self.charity!).getCapability<&{FungibleToken.Receiver}>(/public/fusdReceiver)
                assert(
                    charityVault.borrow() != nil,
                    message: "Could not borrow charity vault"
                )
                charityVault.borrow()!.deposit(from: <-buyerDonation)
            } else {
                sellerVaultCapability.deposit(
                    from: <-buyerDonation
                )
            }

            let nft <- self.sellerItemProvider.borrow()!.withdraw(withdrawID: self.itemID)
            buyerCollection.deposit(token: <-nft)

            emit SaleOfferAccepted(itemID: self.itemID)
        }

        pub fun exchangeNFT() {
            let nft <- self.sellerItemProvider.borrow()!.withdraw(withdrawID: self.itemID)
            self.recipientCollection!.borrow()!.deposit(token: <-nft)
        }

        // destructor
        //
        destroy() {
            // Whether the sale completed or not, publicize that it is being withdrawn.

            emit SaleOfferFinished(itemID: self.itemID)
        }

        // initializer
        // Take the information required to create a sale offer, notably the capability
        // to transfer the NFT and the capability to receive FUSD in payment.
        //
        init(
            sellerItemProvider: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>,
            collection: Address,
            itemID: UInt64,
            typeID: UInt64,
            sellerPaymentReceiver: Capability<&{FungibleToken.Receiver}>,
            platformVaultCapability: Capability<&{FungibleToken.Receiver}>,
            creatorVaultCapability: Capability<&{FungibleToken.Receiver}>,
            communityVaultCapability: Capability<&{FungibleToken.Receiver}>,
            price: UFix64,
            platformFee: UFix64,
            creatorFee: UFix64,
            communityFee: UFix64,
            charity: Address?,
            charityPercentage: UFix64,
            donationAmount: UFix64,
        ) {
            pre {
                sellerItemProvider.borrow() != nil: "Cannot borrow seller."
                sellerPaymentReceiver.borrow() != nil: "Cannot borrow sellerPaymentReceiver."
                // charity.borrow() != nil: "Cannot borrow charity"
                // TODO: Should we change the data type from Address to
                // TODO: Capability<&{FungibleToken.Receiver}> ?
            }

            self.saleCompleted = false

            let collectionRef = sellerItemProvider.borrow()!
            assert(
                collectionRef.borrowNFT(id: itemID) != nil,
                message: "Specified NFT is not available in the owner's collection"
            )

            let collectionRefNFT = collectionRef.borrowNFT(id: itemID) as &NonFungibleToken.NFT

            self.sellerItemProvider = sellerItemProvider
            self.collection = collection
            self.itemID = itemID

            self.sellerPaymentReceiver = sellerPaymentReceiver
            self.platformVaultCapability = platformVaultCapability
            self.creatorVaultCapability = creatorVaultCapability
            self.communityVaultCapability = communityVaultCapability
            self.price = price
            self.platformFee = platformFee
            self.creatorFee = creatorFee
            self.communityFee = communityFee
            self.typeID = typeID

            self.recipientVaultCap = nil
            self.recipientCollection = nil
            self.charity = charity
            self.charityPercentage = charityPercentage
            self.donationAmount = donationAmount

            emit SaleOfferCreated(itemID: self.itemID, collection: self.collection, price: self.price)
        }
    }

    // createSaleOffer
    // Make creating a SaleOffer publicly accessible.
    //
    pub fun createSaleOffer (
        sellerItemProvider: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>,
        collectionPublicPath: PublicPath,
        collection: Address,
        itemID: UInt64,
        typeID: UInt64,
        sellerPaymentReceiver: Capability<&{FungibleToken.Receiver}>,
        price: UFix64,
        charity: Address?,
        charityPercentage: UFix64,
        donationAmount: UFix64,
    ): @SaleOffer {
        let collectionRef = sellerItemProvider.borrow()!
        let nft = collectionRef.borrowNFT(id: itemID) as &NonFungibleToken.NFT
        let metadata = nft.metadata
        let creatorAddress = metadata["creator"] as! Address? ?? panic("No creator address found!")
        let communityAddress = metadata["community"] as! Address? ?? panic("No creator address found!")

        let fullItemId = collection.toString().concat("/".concat(itemID.toString()))
        let platformFee = price * (metadata["platformFee2ndPercentage"] as! UFix64? ?? 0.025)
        let creatorFee = price * (metadata["creatorFee2ndPercentage"] as! UFix64? ?? 0.02)
        let communityFee = price * (metadata["communityFee2ndPercentage"] as! UFix64? ?? 0.03)

        let creatorVaultCapability = getAccount(creatorAddress)
            .getCapability<&{FungibleToken.Receiver}>(/public/fusdReceiver)
        let communityVaultCapability = getAccount(communityAddress)
            .getCapability<&{FungibleToken.Receiver}>(/public/fusdReceiver)

        return <-create SaleOffer(
            sellerItemProvider: sellerItemProvider,
            collection: collection,
            itemID: itemID,
            typeID: typeID,
            sellerPaymentReceiver: sellerPaymentReceiver,
            platformVaultCapability: self.platformVaultCapability,
            creatorVaultCapability: creatorVaultCapability,
            communityVaultCapability: communityVaultCapability,
            price: price,
            platformFee: platformFee,
            creatorFee: creatorFee,
            communityFee: communityFee,
            charity: charity,
            charityPercentage: charityPercentage,
            donationAmount: donationAmount,
        )
    }

    // CollectionManager
    // An interface for adding and removing SaleOffers to a collection, intended for
    // use by the collection's owner.
    //
    pub resource interface CollectionManager {
        pub fun insert(offer: @MetagoodMarket.SaleOffer)
        pub fun remove(collection: Address, itemID: UInt64): @SaleOffer
    }

    // CollectionPublic
    // An interface to allow listing and borrowing SaleOffers, and purchasing items via SaleOffers in a collection.
    //
    pub resource interface CollectionPublic {
        pub fun getSaleOfferIDs(): [String]
        pub fun borrowSaleItem(collection: Address, itemID: UInt64): &SaleOffer{SaleOfferPublicView}?
        pub fun purchase(
            collection: Address,
            itemID: UInt64,
            buyerCollection: &{NonFungibleToken.Receiver},
            buyerPayment: @FungibleToken.Vault,
            buyerDonation: @FungibleToken.Vault,
            buyerPlatformFee: @FungibleToken.Vault,
            buyerCreatorFee: @FungibleToken.Vault,
            buyerCommunityFee: @FungibleToken.Vault,
        )
   }

    // Collection
    // A resource that allows its owner to manage a list of SaleOffers, and purchasers to interact with them.
    //
    pub resource Collection : CollectionManager, CollectionPublic {
        access(self) var saleOffers: @{String: SaleOffer}

        // insert
        // Insert a SaleOffer into the collection, replacing one with the same itemID if present.
        //
         pub fun insert(offer: @MetagoodMarket.SaleOffer) {
            let itemID: UInt64 = offer.itemID
            let collection: Address = offer.collection
            let typeID: UInt64 = offer.typeID
            let price: UFix64 = offer.price

            // add the new offer to the dictionary which removes the old one
            let fullItemId = collection.toString().concat("/".concat(itemID.toString()))
            let oldOffer <- self.saleOffers[fullItemId] <- offer
            destroy oldOffer

            emit CollectionInsertedSaleOffer(
              itemID: itemID,
              collection: collection,
              typeID: typeID,
              owner: self.owner?.address!,
              price: price,
            )
        }

        // remove
        // Remove and return a SaleOffer from the collection.
        pub fun remove(collection: Address, itemID: UInt64): @SaleOffer {
            let fullItemId = collection.toString().concat("/".concat(itemID.toString()))
            emit CollectionRemovedSaleOffer(collection: collection, itemID: itemID, owner: self.owner?.address!)
            return <-(self.saleOffers.remove(key: fullItemId) ?? panic("missing SaleOffer"))
        }

        // purchase
        // If the caller passes a valid itemID and the item is still for sale, and passes a FUSD vault
        // typed as a FungibleToken.Vault (.deposit() handles the type safety of this)
        // containing the correct payment amount, this will transfer the NFT to the caller's
        // collection.
        // It will then remove and destroy the offer.
        // Note that is means that events will be emitted in this order:
        //   1. Collection.CollectionRemovedSaleOffer
        //   2. Collection.Withdraw
        //   3. Collection.Deposit
        //   4. SaleOffer.SaleOfferFinished
        //
        pub fun purchase(
            collection: Address,
            itemID: UInt64,
            buyerCollection: &{NonFungibleToken.Receiver},
            buyerPayment: @FungibleToken.Vault,
            buyerDonation: @FungibleToken.Vault,
            buyerPlatformFee: @FungibleToken.Vault,
            buyerCreatorFee: @FungibleToken.Vault,
            buyerCommunityFee: @FungibleToken.Vault,
        ) {
            pre {
                self.saleOffers[collection.toString().concat("/".concat(itemID.toString()))] != nil: "This sale offer does not exist in the collection!"
            }
            let offer <- self.remove(collection: collection, itemID: itemID)
            let donationAmount = buyerDonation.balance
            offer.accept(
                buyerCollection: buyerCollection,
                buyerPayment: <-buyerPayment,
                buyerDonation: <-buyerDonation,
                buyerPlatformFee: <-buyerPlatformFee,
                buyerCreatorFee: <-buyerCreatorFee,
                buyerCommunityFee: <-buyerCommunityFee,
            )

            emit SaleOfferPurchased(
                itemID: itemID,
                collection: offer.collection,
                from_: self.owner?.address!,
                to_: buyerCollection.owner?.address!,
                price: offer.price,
            )

            emit SaleOfferPlatformFeePaid(
                itemID: itemID,
                collection: offer.collection,
                from_: self.owner?.address!,
                platformFee: offer.platformFee,
            )

            if offer.charity != nil {
                emit SaleOfferDonated(
                    itemID: itemID,
                    from_: self.owner?.address!,
                    to_: offer.charity,
                    amount: donationAmount
                )
            }

            let fullItemId = collection.toString().concat("/".concat(itemID.toString()))

            //FIXME: Is this correct? Or should we return it to the caller to dispose of?
            destroy offer
        }

        // getSaleOfferIDs
        // Returns an array of the IDs that are in the collection
        //
        pub fun getSaleOfferIDs(): [String] {
            return self.saleOffers.keys
        }

        // borrowSaleItem
        // Returns an Optional read-only view of the SaleItem for the given itemID if it is contained by this collection.
        // The optional will be nil if the provided itemID is not present in the collection.
        //
        pub fun borrowSaleItem(collection: Address, itemID: UInt64): &SaleOffer{SaleOfferPublicView}? {
            if self.saleOffers[collection.toString().concat("/".concat(itemID.toString()))] == nil {
                return nil
            }
            return &self.saleOffers[collection.toString().concat("/".concat(itemID.toString()))] as &SaleOffer{SaleOfferPublicView}
        }

        // destructor
        //
        destroy () {
            destroy self.saleOffers
        }

        // constructor
        //
        init () {
            self.saleOffers <- {}
        }
    }

    // createEmptyCollection
    // Make creating a Collection publicly accessible.
    //
    pub fun createEmptyCollection(): @Collection {
        return <-create Collection()
    }

    init () {
        //FIXME: REMOVE SUFFIX BEFORE RELEASE
        self.CollectionStoragePath = /storage/metagoodMarketCollection007
        self.CollectionPublicPath = /public/metagoodMarketCollection007
        self.platformVaultCapability = self.account.getCapability<&{FungibleToken.Receiver}>(/public/fusdReceiver)
    }
}
