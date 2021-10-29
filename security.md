## 1. Metadata ***(NEED GUIDANCE)***

We implemented our own `metadata` field in the `NonFungibleToken` standard token. We followed the exact implementation of `id` in that contract. According to the suggested change,

```bash
pub let metadata: {String: AnyStruct}
```

*Declaring metadata as public allows token owner to modify the contents arbitrarily. If this isn't intended, consider declaring it as private such as `access(self) let metadata: {String: AnyStruct}`*

Doesn't that mean the id should also be `access(self)`? The id specified by Dapper in the NonFungibleToken contract is `pub let`, so does that mean the token owner can modify the id?

Also, changing the metadata to `access(self)` throws this error:

```
error: invalid access modifier for field: `priv`. private fields can never be used
  --> ad5215d0b261af0e.NonFungibleToken:74:8
   |
74 |         access(self) let metadata: {String: AnyStruct}
   |         ^^^^
```

---

## 2. Precondition

```bash
if (MetaBear.totalSupply == self.maxSupply) {
	panic("Max supply reached")
}
```

*Consider using Cadence function [precondition](https://docs.onflow.org/cadence/language/functions/#function-preconditions-and-postconditions), same goes for [L191-201](https://github.com/nfinita/contracts/blob/main/contracts/MetaBear.cdc#L191-L201)*

## Response

Changed the supply check, but could not change L191-201 to precondition because we need to access a resource variable to do some of those checks.

---

## 3. Minter

```bash
self.MinterPublicPath = /public/MetaBearMinter004
```

*Creating a [PublicPath](https://github.com/nfinita/contracts/blob/main/contracts/MetaBear.cdc#L359-L361) for Minter allows anyone to [borrow](https://docs.onflow.org/cadence/anti-patterns/#public-capability-fields-are-a-security-hole) and [mint](https://github.com/nfinita/contracts/blob/main/contracts/MetaBear.cdc#L166) NFTs arbitrarily, is this intended? Theoretically, an attacker can [cause](https://github.com/nfinita/contracts/blob/main/contracts/MetaBear.cdc#L172-L174) mintNFT to panic by minting up to [maxSupply](https://github.com/nfinita/contracts/blob/main/contracts/MetaBear.cdc#L328). Or perhaps you want to restrict specific people to mint by changing into PrivatePath?*

## Response

Yes, this is intentional because our `mintNFT` function only goes through if a payment is included with the function call.

---

## 4. Profile

We removed Profile contract entirely.

## 5. NfinitaMarket

We renamed NfinitaMarket to MetagoodMarket in accordance to our Nfinita's name change to Metagood.

## 6. Bids

```bash
pub let bids: [{Address: SaleOfferBid}]
pub var saleOffers: @{String: SaleOffer}
```
Consider modifying into private access, such as access(self)

# Response
We remove bids from NfinitaMarket for now. Sale Offers are now access(self).

## 7. NfinitaMarket interfaces

```bash
pub resource interface CollectionPurchaser
pub resource interface CollectionPublic
```
Both interfaces hold the same functions which are purchase, placeBid, and settleAuction. Since CollectionPublic will be used in transactions, consider removing duplicate functions(excluding settleAuction, see below) from CollectionPurchaser

# Response
We remove CollectionPurchaser as there is no need for it.

## 8. Settle Auction

```bash
pub fun settleAuction(
    collection: Address,
    itemID: UInt64
)
```
settleAuction is placed inside CollectionPublic interface which means a malicious bidder can close the auction instantly after placing their bid, resulting in unfair bidding experience for other bidders and loss of profit for seller. Consider removing settleAuction from CollectionPublic and place it in CollectionPurchaser interface so the public capability doesn't expose it.

```bash
// Expiry Date in seconds, unix time
pub let expiryDate: UInt64
```

From the comment and variable name, it looks like it's an expiration date for bidding (value other than 0 is Auction only). However, there's no expiration date check when placing bids, is this intended?
It seems like it's possible to place bid for a completed sale offer, consider adding precondition to only able place bids if the sale haven't been completed yet(also goes for settleAuction)

A malicious seller can wait for a user to place bid, then calls settleBids to drain bidder's funds to seller's wallet. Setting sale completed to true does not prevents it since other bidders can still place bids. In result, the outbidded bidder would not have their funds returned by the vault.
These public functions can be directly called by the seller, consider adding access(contract) to make sure it's only controlled by Collection resource.

# Response
We remove bids from NfinitaMarket for now, which includes expiryDate and settleAuction()
