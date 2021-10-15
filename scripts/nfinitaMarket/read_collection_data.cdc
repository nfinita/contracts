// TEST PURPOSE
import MetaBear from "../../contracts/MetaBear.cdc"

pub fun main(address: Address): [AnyStruct] {
    let ref = getAccount(address)
        .getCapability<&{MetaBear.CollectionDataPublic}>(
            MetaBear.CollectionDataPublicPath
        )
        .borrow()
        ?? panic("Could not borrow market collection from market address")

    return ref.getCollectionMetadata()
}
