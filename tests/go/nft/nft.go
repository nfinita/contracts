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
	// nftcontracts "github.com/onflow/flow-nft/lib/go/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nfinita/first-market/cadence/tests/go/test"
)

const (
	metaBearTransactionsRootPath = "../../transactions/metaBear"
	metaBearScriptsRootPath      = "../../scripts/metaBear"

	metaBearContractPath         = "../../contracts/MetaBear.cdc"
	nonFungibleTokenContractPath = "../../contracts/NonFungibleToken.cdc"

	metaBearSetupAccountPath          = metaBearTransactionsRootPath + "/setup_meta_bear.cdc"
	metaBearSetupAccountMetaBearPath  = metaBearTransactionsRootPath + "/setup_account_meta_bear.cdc"
	metaBearMintMetaBearPath          = metaBearTransactionsRootPath + "/mint_meta_bear.cdc"
	metaBearTransferMetaBearPath      = metaBearTransactionsRootPath + "/transfer_meta_bear.cdc"
	metaBearSetupMetaBearSettingsPath = metaBearTransactionsRootPath + "/set_metabear_settings.cdc"
	metaBearGetMetaBearSupplyPath     = metaBearScriptsRootPath + "/get_meta_bear_supply.cdc"
	metaBearGetCollectionLengthPath   = metaBearScriptsRootPath + "/get_collection_length.cdc"

	metaBearMetaDataPath = "../../dummydata/metaBear/metadata.json"
)

func DeployContracts(
	t *testing.T,
	b *emulator.Blockchain,
) (flow.Address, flow.Address, flow.Address, flow.Address, crypto.Signer, crypto.Signer) {
	accountKeys := sdktest.AccountKeyGenerator()

	// should be able to deploy a contract as a new account with no keys
	// nftCode := nftcontracts.NonFungibleToken()
	nftCode := loadNonFungibleToken()
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
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// simplify the workflow by having the contract address also be our initial test collection
	SetupAccount(t, b, metaBearAddr, metaBearSigner, nftAddress, metaBearAddr)

	return ftAddress, fusdAddr, nftAddress, metaBearAddr, metaBearSigner, fusdSigner
}

func SetupAccountMetaBear(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress flow.Address,
	userSigner crypto.Signer,
	nftAddress flow.Address,
	metaBearAddress flow.Address,
) {
	tx := flow.NewTransaction().
		SetScript(SetupAccountMetaBearScript(nftAddress.String(), metaBearAddress.String())).
		SetGasLimit(1000).
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
	nftAddress, metaBearAddr, ftAddress, fusdAddress, userAddress flow.Address,
	userSigner crypto.Signer,
) {
	tx := flow.NewTransaction().
		SetScript(MintMetaBearScript(nftAddress.String(), metaBearAddr.String(), ftAddress.String(), fusdAddress.String())).
		SetGasLimit(1000).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewAddress(metaBearAddr))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		false,
	)
}

func TransferItem(
	t *testing.T, b *emulator.Blockchain,
	nftAddress, metaBearAddr, userAddr flow.Address, userSigner crypto.Signer,
	typeID uint64, recipientAddr flow.Address, shouldFail bool,
) {

	tx := flow.NewTransaction().
		SetScript(TransferMetaBearScript(nftAddress.String(), metaBearAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddr)

	_ = tx.AddArgument(cadence.NewAddress(recipientAddr))
	_ = tx.AddArgument(cadence.NewUInt64(typeID))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func SetMetaBearSettings(
	t *testing.T, b *emulator.Blockchain,
	nftAddress, metaBearAddr flow.Address,
	metaBearSigner crypto.Signer,
	community flow.Address,
	communityFee2ndPercentage string,
	creator flow.Address,
	creatorFeeMintPercentage string,
	creatorFee2ndPercentage string,
	platform flow.Address,
	platformFeeMintPercentage string,
	platformFee2ndPercentage string,
	shouldFail bool,
) {

	tx := flow.NewTransaction().
		SetScript(SetMetaBearSettingsScript(nftAddress.String(), metaBearAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(metaBearAddr)

	_ = tx.AddArgument(cadence.NewAddress(community))
	_ = tx.AddArgument(test.CadenceUFix64(communityFee2ndPercentage))
	_ = tx.AddArgument(cadence.NewAddress(creator))
	_ = tx.AddArgument(test.CadenceUFix64(creatorFeeMintPercentage))
	_ = tx.AddArgument(test.CadenceUFix64(creatorFee2ndPercentage))
	_ = tx.AddArgument(cadence.NewAddress(platform))
	_ = tx.AddArgument(test.CadenceUFix64(platformFeeMintPercentage))
	_ = tx.AddArgument(test.CadenceUFix64(platformFee2ndPercentage))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, metaBearAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), metaBearSigner},
		shouldFail,
	)
}

func loadNonFungibleToken() []byte {
	return test.ReadFile(nonFungibleTokenContractPath)
}

func loadMetaBear(nftAddress string, ftAddress string, fusdAddress string) []byte {
	return []byte(test.ReplaceImports(
		string(test.ReadFile(metaBearContractPath)),
		map[string]*regexp.Regexp{
			nftAddress:  test.NonFungibleTokenAddressPlaceholder,
			ftAddress:   test.FungibleTokenAddressPlaceholder,
			fusdAddress: test.FUSDAddressPlaceHolder,
		},
	))
}

func replaceAddressPlaceholders(code, nftAddress, metaBearAddress, ftAddress, fusdAddress, metadata string) []byte {
	/*
		print("\nNFT CODE ====\n")
		print(test.ReplaceEnvs(test.ReplaceImports(
			code,
			map[string]*regexp.Regexp{
				nftAddress:      test.NonFungibleTokenAddressPlaceholder,
				metaBearAddress: test.MetaBearAddressPlaceHolder,
				ftAddress:       test.FungibleTokenAddressPlaceholder,
				fusdAddress:     test.FUSDAddressPlaceHolder,
			},
		), map[string]*regexp.Regexp{
			metadata: test.MetadataPlaceholder,
		}))
		print("\nNFT CODE END ====\n")
	*/
	return []byte(test.ReplaceEnvs(test.ReplaceImports(
		code,
		map[string]*regexp.Regexp{
			nftAddress:      test.NonFungibleTokenAddressPlaceholder,
			metaBearAddress: test.MetaBearAddressPlaceHolder,
			ftAddress:       test.FungibleTokenAddressPlaceholder,
			fusdAddress:     test.FUSDAddressPlaceHolder,
		},
	), map[string]*regexp.Regexp{
		metadata: test.MetadataPlaceholder,
	}))
}

func SetupAccountMetaBearScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearSetupAccountMetaBearPath)),
		nftAddress,
		metaBearAddress,
		"",
		"",
		"",
	)
}

func SetupAccountScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearSetupAccountPath)),
		nftAddress,
		metaBearAddress,
		"",
		"",
		string(test.ReadFile(metaBearMetaDataPath)),
	)
}

func MintMetaBearScript(nftAddress, metaBearAddress, ftAddress, fusdAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearMintMetaBearPath)),
		nftAddress,
		metaBearAddress,
		ftAddress,
		fusdAddress,
		"",
	)
}

func TransferMetaBearScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearTransferMetaBearPath)),
		nftAddress,
		metaBearAddress,
		"",
		"",
		"",
	)
}

func SetMetaBearSettingsScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearSetupMetaBearSettingsPath)),
		nftAddress,
		metaBearAddress,
		"",
		"",
		"",
	)
}

func GetMetaBearSupplyScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearGetMetaBearSupplyPath)),
		nftAddress,
		metaBearAddress,
		"",
		"",
		"",
	)
}

func GetCollectionLengthScript(nftAddress, metaBearAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(metaBearGetCollectionLengthPath)),
		nftAddress,
		metaBearAddress,
		"",
		"",
		"",
	)
}
