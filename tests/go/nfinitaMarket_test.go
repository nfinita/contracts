package test

import (
	"github.com/nfinita/first-market/cadence/tests/go/fusd"
	"testing"

	"github.com/nfinita/first-market/cadence/tests/go/nfinitaMarket"
	"github.com/nfinita/first-market/cadence/tests/go/nft"
	"github.com/nfinita/first-market/cadence/tests/go/test"
)

func TestNfinitaMarketDeployContracts(t *testing.T) {
	b := test.NewBlockchain()

	nfinitaMarket.DeployContracts(t, b)
}

func TestNfinitaMarketSetupAccount(t *testing.T) {
	b := test.NewBlockchain()

	contracts := nfinitaMarket.DeployContracts(t, b)

	t.Run("Should be able to create an empty Storefront", func(t *testing.T) {
		userAddress, userSigner, _ := test.CreateAccount(t, b)
		nfinitaMarket.SetupAccount(t, b, userAddress, userSigner, contracts.NfinitaMarketAddress)
	})
}

func TestNfinitaMarketCreateSaleOffer(t *testing.T) {
	b := test.NewBlockchain()

	contracts := nfinitaMarket.DeployContracts(t, b)

	t.Run("Should be able to create a sale offer and list it", func(t *testing.T) {
		tokenToList := uint64(0)
		tokenPrice := "1.11"
		userAddress, userSigner := nfinitaMarket.CreateAccount(t, b, contracts)

		// contract mints item
		nft.MintItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.MetaBearSigner,
			typeID,
		)

		// contract transfers item to another seller account (we don't need to do this)
		nft.TransferItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.MetaBearSigner,
			tokenToList,
			userAddress,
			false,
		)

		// other seller account lists the item
		nfinitaMarket.ListItem(
			t, b,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
	})

	t.Run("Should be able to accept a sale offer", func(t *testing.T) {
		tokenToList := uint64(1)
		tokenPrice := "1.11"
		userAddress, userSigner := nfinitaMarket.CreateAccount(t, b, contracts)

		// contract mints item
		nft.MintItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.MetaBearSigner,
			typeID,
		)

		// contract transfers item to another seller account (we don't need to do this)
		nft.TransferItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.MetaBearSigner,
			tokenToList,
			userAddress,
			false,
		)

		// other seller account lists the item
		saleOfferResourceID := nfinitaMarket.ListItem(
			t, b,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)

		buyerAddress, buyerSigner := nfinitaMarket.CreatePurchaserAccount(t, b, contracts)

		// fund the purchase
		fusd.Mint(
			t, b,
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
			contracts.FUSDSigner,
			buyerAddress,
			"100.0",
			false,
		)

		// Make the purchase
		nfinitaMarket.PurchaseItem(
			t, b,
			contracts,
			buyerAddress,
			buyerSigner,
			userAddress,
			saleOfferResourceID,
			false,
		)
	})

	t.Run("Should be able to remove a sale offer", func(t *testing.T) {
		tokenToList := uint64(2)
		tokenPrice := "1.11"
		userAddress, userSigner := nfinitaMarket.CreateAccount(t, b, contracts)

		// contract mints item
		nft.MintItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.MetaBearSigner,
			typeID,
		)

		// contract transfers item to another seller account (we don't need to do this)
		nft.TransferItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.MetaBearSigner,
			tokenToList,
			userAddress,
			false,
		)

		// other seller account lists the item
		saleOfferResourceID := nfinitaMarket.ListItem(
			t, b,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)

		// make the purchase
		nfinitaMarket.RemoveItem(
			t, b,
			contracts,
			userAddress,
			userSigner,
			saleOfferResourceID,
			false,
		)
	})
}
