package test

import (
	"github.com/nfinita/first-market/cadence/tests/go/fusd"
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/stretchr/testify/assert"

	"github.com/nfinita/first-market/cadence/tests/go/nft"
	"github.com/nfinita/first-market/cadence/tests/go/test"
)

const typeID = 1000

func TestMetaBearDeployContracts(t *testing.T) {
	b := test.NewBlockchain()
	nft.DeployContracts(t, b)
}

func TestCreateMetaBear(t *testing.T) {
	b := test.NewBlockchain()

	ftAddress, fusdAddress, nftAddress, metaBearAddr, metaBearSigner, fusdSigner :=
		nft.DeployContracts(t, b)

	userAddress, userSigner, _ := test.CreateAccount(t, b)
	nft.SetupAccountMetaBear(t, b, userAddress, userSigner, nftAddress, metaBearAddr)
	nft.SetMetaBearSettings(
		t,
		b,
		nftAddress,
		metaBearAddr,
		metaBearSigner,
		userAddress, // TODO: Change to another account
		"0.03",
		userAddress, // TODO: Change to another account
		"0.08",
		"0.02",
		userAddress, // TODO: Change to another account
		"0.05",
		"0.025",
		false,
	)

	// fund the mint
	fusd.SetupAccount(
		t, b,
		userAddress,
		userSigner,
		ftAddress,
		fusdAddress,
	)
	fusd.Mint(
		t, b,
		ftAddress,
		fusdAddress,
		fusdSigner,
		userAddress,
		"100.0",
		false,
	)

	supply := test.ExecuteScriptAndCheck(
		t, b,
		nft.GetMetaBearSupplyScript(nftAddress.String(), metaBearAddr.String()),
		nil,
	)
	assert.EqualValues(t, cadence.NewUInt64(0), supply)

	// assert that the account collection is empty
	length := test.ExecuteScriptAndCheck(
		t, b,
		nft.GetCollectionLengthScript(nftAddress.String(), metaBearAddr.String()),
		[][]byte{jsoncdc.MustEncode(cadence.NewAddress(userAddress))},
	)
	assert.EqualValues(t, cadence.NewInt(0), length)

	t.Run("Should be able to mint a metaBear", func(t *testing.T) {
		nft.MintItem(t, b, nftAddress, metaBearAddr, ftAddress, fusdAddress, userAddress, userSigner)

		// assert that the account collection is correct length
		length = test.ExecuteScriptAndCheck(
			t, b,
			nft.GetCollectionLengthScript(nftAddress.String(), metaBearAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(userAddress))},
		)
		assert.EqualValues(t, cadence.NewInt(1), length)
	})
}

func TestTransferNFT(t *testing.T) {
	b := test.NewBlockchain()

	ftAddress, fusdAddress, nftAddress, metaBearAddr, metaBearSigner, fusdSigner :=
		nft.DeployContracts(t, b)

	userAddress, userSigner, _ := test.CreateAccount(t, b)
	nft.SetupAccountMetaBear(t, b, userAddress, userSigner, nftAddress, metaBearAddr)
	nft.SetMetaBearSettings(
		t,
		b,
		nftAddress,
		metaBearAddr,
		metaBearSigner,
		userAddress, // TODO: Change to another account
		"0.03",
		userAddress, // TODO: Change to another account
		"0.08",
		"0.02",
		userAddress, // TODO: Change to another account
		"0.05",
		"0.025",
		false,
	)

	// create a new Collection
	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {
		nft.SetupAccount(t, b, metaBearAddr, metaBearSigner, nftAddress, metaBearAddr)

		length := test.ExecuteScriptAndCheck(
			t, b,
			nft.GetCollectionLengthScript(nftAddress.String(), metaBearAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(userAddress))},
		)
		assert.EqualValues(t, cadence.NewInt(0), length)
	})

	t.Run("Should not be able to withdraw an NFT that does not exist in a collection", func(t *testing.T) {
		nonExistentID := uint64(3333333)

		nft.TransferItem(
			t, b,
			nftAddress, metaBearAddr, userAddress, userSigner,
			nonExistentID, userAddress, true,
		)
	})

	// transfer an NFT
	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		fusd.SetupAccount(t, b, userAddress, userSigner, ftAddress, fusdAddress)
		fusd.Mint(t, b, ftAddress, fusdAddress, fusdSigner, userAddress, "100.00", false)
		nft.MintItem(t, b, nftAddress, metaBearAddr, ftAddress, fusdAddress, userAddress, userSigner)
		// Cheat: we have minted one item, its ID will be zero
		nft.TransferItem(
			t, b,
			nftAddress, metaBearAddr, userAddress, userSigner,
			0, userAddress,  false,
		)
	})
}
