import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MetaBear from "../../contracts/MetaBear.cdc"

pub fun hasCollection(_ address: Address): Bool {
  return getAccount(address)
    .getCapability<&MetaBear.Collection{NonFungibleToken.CollectionPublic, MetaBear.MetaBearCollectionPublic}>(
        MetaBear.CollectionPublicPath
    ).check()
}

pub fun main(address: Address): UInt64 {
    return hasCollection(address)
}
