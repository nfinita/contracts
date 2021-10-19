import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MetaBear from "../../contracts/MetaBear.cdc"

// This script returns the size of an account's MetaBear collection.

pub fun main(address: Address): Int {
    let account = getAccount(address)

    let collectionRef = account.getCapability(MetaBear.CollectionPublicPath)!
        .borrow<&{NonFungibleToken.CollectionPublic}>()
        ?? panic("Could not borrow capability from public collection")

    return collectionRef.getIDs().length
}
