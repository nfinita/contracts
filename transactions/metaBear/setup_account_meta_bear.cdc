import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MetaBear from "../../contracts/MetaBear.cdc"

// This transction uses the NFTMinter resource to mint a new NFT.
//
// It must be run with the account that has the minter resource
// stored at path /storage/NFTMinter.

pub fun hasCollection(_ address: Address): Bool {
  return getAccount(address)
    .getCapability<&MetaBear.Collection{NonFungibleToken.CollectionPublic, MetaBear.MetaBearCollectionPublic}>(MetaBear.CollectionPublicPath)
    .check()
}

transaction {
    prepare(signer: AuthAccount) {
        if !hasCollection(signer.address) {
            if signer.borrow<&MetaBear.Collection>(from: MetaBear.CollectionStoragePath) == nil {
                signer.save(<-MetaBear.createEmptyCollection(), to: MetaBear.CollectionStoragePath)
            }
            signer.unlink(MetaBear.CollectionPublicPath)
            signer.link<&MetaBear.Collection{NonFungibleToken.CollectionPublic, MetaBear.MetaBearCollectionPublic}>(MetaBear.CollectionPublicPath, target: MetaBear.CollectionStoragePath)
        }
    }

    execute {

    }
}
