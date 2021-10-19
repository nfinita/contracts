package test

import (
	"github.com/nfinita/first-market/cadence/tests/go/fusd"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
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
		nfinitaMarket.SetupAccount(
			t,
			b,
			userAddress,
			userSigner,
			contracts.NfinitaMarketAddress,
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
		)
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
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
			userAddress,
			userSigner,
		)

		/*
		// contract transfers item to another seller account (we don't need to do this)
		nft.TransferItem(
			t, b,
			contracts.NonFungibleTokenAddress,
		    contracts.MetaBearAddress,
			userAddress,
			userSigner,
			tokenToList,
			userAddress,
			false,
		)
		*/

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
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
			userAddress,
			userSigner,
		)

		// contract transfers item to another seller account (we don't need to do this)
		nft.TransferItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			userAddress,
			userSigner,
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
		nft.SetupAccountMetaBear(
			t, b,
			buyerAddress,
			buyerSigner,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
		 )

		supply := test.ExecuteScriptAndCheck(
			t, b,
			nfinitaMarket.ReadCollectionIdsScript(contracts),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(userAddress))},
		)

		print("Sale Offer Resource ID\n")
		print(saleOfferResourceID)
		print("\n")
		print(contracts.MetaBearAddress.String())
		print("\n")
		print(supply.String())
		print("\nSale Offer Resource ID END\n")

		// Make the purchase
		nfinitaMarket.PurchaseItem(
			t, b,
			contracts,
			buyerAddress,
			buyerSigner,
			saleOfferResourceID,
			"0.0",
			userAddress,
			false,
		)
	})

	t.Run("Should be able to remove a sale offer", func(t *testing.T) {
		tokenToList := uint64(2)
		tokenPrice := "1.11"
		userAddress, userSigner := nfinitaMarket.CreateAccount(t, b, contracts)

		// fund the mint
		fusd.SetupAccount(
			t, b,
			userAddress,
			userSigner,
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
		)
		fusd.Mint(
			t, b,
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
			contracts.FUSDSigner,
			userAddress,
			"100.0",
			false,
		)

		// contract mints item
		nft.MintItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			contracts.FungibleTokenAddress,
			contracts.FUSDAddress,
			userAddress,
			userSigner,
		)

		// contract transfers item to another seller account (we don't need to do this)
		nft.TransferItem(
			t, b,
			contracts.NonFungibleTokenAddress,
			contracts.MetaBearAddress,
			userAddress,
			userSigner,
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
