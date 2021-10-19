import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import FungibleToken from "../../contracts/FungibleToken.cdc"
import MetaBear from "../../contracts/MetaBear.cdc"
import FUSD from "../../contracts/FUSD.cdc"

// This transction uses the NFTMinter resource to mint a new NFT.
//
// It must be run with the account that has the minter resource
// stored at path /storage/NFTMinter.

pub fun hasCollection(_ address: Address): Bool {
  return getAccount(address)
    .getCapability<&MetaBear.Collection{NonFungibleToken.CollectionPublic, MetaBear.MetaBearCollectionPublic}>(MetaBear.CollectionPublicPath)
    .check()
}

transaction(collection: Address) {
    // local variable for storing the minter reference
    var minter: &MetaBear.NFTMinter

    // The Vault resource that holds the tokens that are being transfered
    let communityFeeVault: @FungibleToken.Vault
    let creatorFeeVault: @FungibleToken.Vault
    let platformFeeVault: @FungibleToken.Vault

    // Receiver for the NFT
    var receiver: &{NonFungibleToken.CollectionPublic}

    prepare(signer: AuthAccount) {
        // If signer does not have a MetaBear collection, create one
        if !hasCollection(signer.address) {
            if signer.borrow<&MetaBear.Collection>(from: MetaBear.CollectionStoragePath) == nil {
                signer.save(<-MetaBear.createEmptyCollection(), to: MetaBear.CollectionStoragePath)
            }
            signer.unlink(MetaBear.CollectionPublicPath)
            signer.link<&MetaBear.Collection{NonFungibleToken.CollectionPublic, MetaBear.MetaBearCollectionPublic}>(MetaBear.CollectionPublicPath, target: MetaBear.CollectionStoragePath)
        }

        self.minter = getAccount(collection).getCapability<&MetaBear.NFTMinter>(MetaBear.MinterPublicPath).borrow()!

        // Get a reference to the signer's stored vault
        let vaultRef = signer.borrow<&FUSD.Vault>(from: /storage/fusdVault)
        ?? panic("Could not borrow reference to the signer's FUSD vault")

        let metaBearData = getAccount(collection)
            .getCapability<&{MetaBear.CollectionDataPublic}>(MetaBear.CollectionDataPublicPath)
            .borrow() ?? panic("Could not borrow meta bear metadata")

        let price = self.minter.mintPrice
        let creatorFee = price * metaBearData.getCollectionSetting("creatorFeeMintPercentage") as! UFix64
        let platformFee = price * metaBearData.getCollectionSetting("platformFeeMintPercentage") as! UFix64

        // Withdraw tokens from the signer's stored vault
        self.communityFeeVault <- vaultRef.withdraw(amount: price - creatorFee - platformFee)
        self.creatorFeeVault <- vaultRef.withdraw(amount: creatorFee)
        self.platformFeeVault <- vaultRef.withdraw(amount: platformFee)

        // borrow the recipient's public NFT collection reference
        self.receiver = signer.getCapability(MetaBear.CollectionPublicPath).borrow<&{NonFungibleToken.CollectionPublic}>()!
    }

    execute {
        self.minter.mintNFT(
            recipient: self.receiver,
            communityFeeVault: <-self.communityFeeVault,
            creatorFeeVault: <-self.creatorFeeVault,
            platformFeeVault: <-self.platformFeeVault,
        )
    }
}
