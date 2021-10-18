package nft

import (
	"github.com/nfinita/first-market/cadence/tests/go/fusd"
	"regexp"
	"testing"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	sdktest "github.com/onflow/flow-go-sdk/test"
	nftcontracts "github.com/onflow/flow-nft/lib/go/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nfinita/first-market/cadence/tests/go/test"
)

const (
	metaBearTransactionsRootPath = "../../transactions/metaBear"
	metaBearScriptsRootPath      = "../../scripts/metaBear"

	metaBearContractPath            = "../../contracts/MetaBear.cdc"
	metaBearSetupAccountPath        = metaBearTransactionsRootPath + "/setup_meta_bear.cdc"
	metaBearMintMetaBearPath       = metaBearTransactionsRootPath + "/mint_meta_bear.cdc"
	metaBearTransferMetaBearPath   = metaBearTransactionsRootPath + "/transfer_meta_bear.cdc"
	metaBearGetMetaBearSupplyPath  = metaBearScriptsRootPath + "/get_meta_bear_supply.cdc"
	metaBearGetCollectionLengthPath = metaBearScriptsRootPath + "/get_collection_length.cdc"
)

func DeployContracts(
	t *testing.T,
	b *emulator.Blockchain,
) (flow.Address, flow.Address, flow.Address, flow.Address, crypto.Signer, crypto.Signer) {
	accountKeys := sdktest.AccountKeyGenerator()

	// should be able to deploy a contract as a new account with no keys
	nftCode := nftcontracts.NonFungibleToken()
	nftAddress, err := b.CreateAccount(
		nil,
		[]sdktemplates.Contract{
			{
				Name:   "NonFungibleToken",
				Source: string(nftCode),
			},
		},
	)
	require.NoError(t, err)

	// should be able to deploy a contract as a new account with no keys
	ftAddress, fusdAddr, fusdSigner := fusd.DeployContracts(t, b)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// should be able to deploy a contract as a new account with one key
	metaBearAccountKey, metaBearSigner := accountKeys.NewWithSigner()
	metaBearCode := loadMetaBear(nftAddress.String(), ftAddress.String(), fusdAddr.String())

	metaBearAddr, err := b.CreateAccount(
		[]*flow.AccountKey{metaBearAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "MetaBear",
				Source: string(metaBearCode),
			},
		},
	)
	print("METABEAR ADDRESS CODE =======\n")
	print(metaBearAccountKey, "\n", metaBearAddr.String())
	print("\nMETABEAR ADDRESS CODE END =======\n")
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// simplify the workflow by having the contract address also be our initial test collection
	SetupAccount(t, b, metaBearAddr, metaBearSigner, nftAddress, metaBearAddr)

	return ftAddress, fusdAddr, nftAddress, metaBearAddr, metaBearSigner, fusdSigner
}

func SetupAccount(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress flow.Address,
	userSigner crypto.Signer,
	nftAddress flow.Address,
	metaBearAddress flow.Address,
) {
	tx := flow.NewTransaction().
		SetScript(SetupAccountScript(nftAddress.String(), metaBearAddress.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		false,
	)
}

func MintItem(
	t *testing.T, b *emulator.Blockchain,
	nftAddress, metaBearAddr flow.Address,
	metaBearSigner crypto.Signer, typeID uint64,
) {
	metadata := []cadence.KeyValuePair{
		{
			Key:   cadence.String("title"),
			Value: cadence.String("New NFT"),
		},
		{
			Key:   cadence.String("description"),
			Value: cadence.String("New NFT Description"),
		},
		{
			Key:   cadence.String("imageURL"),
			Value: cadence.String("https://some.url/image"),
		},
	}
	tx := flow.NewTransaction().
		SetScript(MintMetaBearScript(nftAddress.String(), metaBearAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(metaBearAddr)

	_ = tx.AddArgument(cadence.NewAddress(metaBearAddr))
	_ = tx.AddArgument(cadence.NewUInt64(typeID))
	_ = tx.AddArgument(cadence.NewDictionary(metadata))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, metaBearAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), metaBearSigner},
		false,
	)
}

func TransferItem(
	t *testing.T, b *emulator.Blockchain,
	nftAddress, metaBearAddr flow.Address, metaBearSigner crypto.Signer,
	typeID uint64, recipientAddr flow.Address, shouldFail bool,
) {

	tx := flow.NewTransaction().
		SetScript(TransferMetaBearScript(nftAddress.String(), metaBearAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(metaBearAddr)

	_ = tx.AddArgument(cadence.NewAddress(recipientAddr))
	_ = tx.AddArgument(cadence.NewUInt64(typeID))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, metaBearAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), metaBearSigner},
		shouldFail,
	)
}

func loadMetaBear(nftAddress string, ftAddress string, fusdAddress string) []byte {
	print("\nMETA BEAR CODE ====\n")
	print(test.ReplaceImports(
		string(test.ReadFile(metaBearContractPath)),
		map[string]*regexp.Regexp{
			nftAddress: test.NonFungibleTokenAddressPlaceholder,
			ftAddress: test.FungibleTokenAddressPlaceholder,
			fusdAddress: test.FUSDAddressPlaceHolder,
		},
	))
	print("\n===== META BEAR CODE END ====\n")
	return []byte(test.ReplaceImports(
		string(test.ReadFile(metaBearContractPath)),
		map[string]*regexp.Regexp{
			nftAddress: test.NonFungibleTokenAddressPlaceholder,
			ftAddress: test.FungibleTokenAddressPlaceholder,
			fusdAddress: test.FUSDAddressPlaceHolder,
		},
	))
}

func replaceAddressPlaceholders(code, nftAddress, metaBearAddress string) []byte {
	return []byte(test.ReplaceImports(
		code,
		map[string]*regexp.Regexp{
			nftAddress:      test.NonFungibleTokenAddressPlaceholder,
			metaBearAddress: test.MetaBearAddressPlaceHolder,
		},
	))
}

func SetupAccountScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearSetupAccountPath)),
		nftAddress,
		metaBearAddress,
	)
}

func MintMetaBearScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearMintMetaBearPath)),
		nftAddress,
		metaBearAddress,
	)
}

func TransferMetaBearScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearTransferMetaBearPath)),
		nftAddress,
		metaBearAddress,
	)
}

func GetMetaBearSupplyScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearGetMetaBearSupplyPath)),
		nftAddress,
		metaBearAddress,
	)
}

func GetCollectionLengthScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearGetCollectionLengthPath)),
		nftAddress,
		metaBearAddress,
	)
}
