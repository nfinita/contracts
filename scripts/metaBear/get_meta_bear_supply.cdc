import MetaBear from "../../contracts/MetaBear.cdc"

// This scripts returns the number of KittyItems currently in existence.

pub fun main(): UInt64 {
    return MetaBear.totalSupply
}