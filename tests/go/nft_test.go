package test

import (
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

	_, _, nftAddress, metaBearAddr, _, metaBearSigner :=
		nft.DeployContracts(t, b)

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
		[][]byte{jsoncdc.MustEncode(cadence.NewAddress(metaBearAddr))},
	)
	assert.EqualValues(t, cadence.NewInt(0), length)

	t.Run("Should be able to mint a metaBear", func(t *testing.T) {
		nft.MintItem(t, b, nftAddress, metaBearAddr, metaBearSigner, typeID)

		// assert that the account collection is correct length
		length = test.ExecuteScriptAndCheck(
			t, b,
			nft.GetCollectionLengthScript(nftAddress.String(), metaBearAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(metaBearAddr))},
		)
		assert.EqualValues(t, cadence.NewInt(1), length)
	})
}

func TestTransferNFT(t *testing.T) {
	b := test.NewBlockchain()

	_, _, nftAddress, metaBearAddr, _, metaBearSigner :=
		nft.DeployContracts(t, b)

	userAddress, userSigner, _ := test.CreateAccount(t, b)

	// create a new Collection
	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {
		nft.SetupAccount(t, b, userAddress, userSigner, nftAddress, metaBearAddr)

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
			nftAddress, metaBearAddr, metaBearSigner,
			nonExistentID, userAddress, true,
		)
	})

	// transfer an NFT
	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		nft.MintItem(t, b, nftAddress, metaBearAddr, metaBearSigner, typeID)
		// Cheat: we have minted one item, its ID will be zero
		nft.TransferItem(
			t, b,
			nftAddress, metaBearAddr, metaBearSigner,
			0, userAddress, false,
		)
	})
}
