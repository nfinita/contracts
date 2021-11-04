1-4 Resolved

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
