import MetagoodMarket from "../../contracts/MetagoodMarket.cdc"
import FungibleToken from "../../contracts/FungibleToken.cdc"
import FUSD from "../../contracts/FUSD.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction {
    prepare(signer: AuthAccount) {

        // if the account doesn't already have a collection
        if signer.borrow<&MetagoodMarket.Collection>(from: MetagoodMarket.CollectionStoragePath) == nil {

            // create a new empty collection
            let collection <- MetagoodMarket.createEmptyCollection() as! @MetagoodMarket.Collection

            // save it to the account
            signer.save(<-collection, to: MetagoodMarket.CollectionStoragePath)

            // create a public capability for the collection
            signer.link<&MetagoodMarket.Collection{MetagoodMarket.CollectionPublic}>(MetagoodMarket.CollectionPublicPath, target: MetagoodMarket.CollectionStoragePath)
        }

        // It's OK if the account already has a Vault, but we don't want to replace it
        if signer.borrow<&FUSD.Vault>(from: /storage/fusdVault) == nil {
            // Create a new FUSD Vault and put it in storage
            signer.save(<-FUSD.createEmptyVault(), to: /storage/fusdVault)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            signer.link<&FUSD.Vault{FungibleToken.Receiver}>(
                /public/fusdReceiver,
                target: /storage/fusdVault
            )

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            signer.link<&FUSD.Vault{FungibleToken.Balance}>(
                /public/fusdBalance,
                target: /storage/fusdVault
            )
        }
    }
}
