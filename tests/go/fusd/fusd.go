package fusd

import (
	"github.com/onflow/cadence"
	"regexp"
	"testing"

	emulator "github.com/onflow/flow-emulator"
	ftcontracts "github.com/onflow/flow-ft/lib/go/contracts"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	sdktest "github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nfinita/first-market/cadence/tests/go/test"
)

const (
	fusdTransactionsRootPath = "../../transactions/fusd"
	fusdScriptsRootPath      = "../../scripts/fusd"
	fusdContractPath         = "../../contracts/FUSD.cdc"
	fusdSetupAccountPath     = fusdTransactionsRootPath + "/setup_account.cdc"
	fusdMintTokensPath       = fusdTransactionsRootPath + "/mint_tokens.cdc"
)

func DeployContracts(
	t *testing.T,
	b *emulator.Blockchain,
) (flow.Address, flow.Address, crypto.Signer) {
	accountKeys := sdktest.AccountKeyGenerator()

	// should be able to deploy a contract as a new account with no keys
	ftCode := ftcontracts.FungibleToken()
	ftAddress, err := b.CreateAccount(
		nil,
		[]sdktemplates.Contract{
			{
				Name:   "FungibleToken",
				Source: string(ftCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// should be able to deploy a contract as a new account with one key
	fusdAccountKey, fusdSigner := accountKeys.NewWithSigner()
	fusdCode := loadFUSD(ftAddress.String())
	fusdAddr, err := b.CreateAccount(
		[]*flow.AccountKey{fusdAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "FUSD",
				Source: string(fusdCode),
			},
		},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// simplify the workflow by having the contract address also be our initial test collection
	SetupAccount(t, b, fusdAddr, fusdSigner, ftAddress, fusdAddr)

	return ftAddress, fusdAddr, fusdSigner
}

func SetupAccount(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress flow.Address,
	userSigner crypto.Signer,
	ftAddress flow.Address,
	fusdAddress flow.Address,
) {
	tx := flow.NewTransaction().
		SetScript(SetupAccountScript(ftAddress.String(), fusdAddress.String())).
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

func Mint(
	t *testing.T,
	b *emulator.Blockchain,
	fungibleTokenAddress flow.Address,
	fusdAddress flow.Address,
	fusdSigner crypto.Signer,
	recipientAddress flow.Address,
	amount string,
	shouldRevert bool,
) {
	tx := flow.NewTransaction().
		SetScript(MintFUSDTransaction(fungibleTokenAddress, fusdAddress)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(fusdAddress)

	_ = tx.AddArgument(cadence.NewAddress(recipientAddress))
	_ = tx.AddArgument(test.CadenceUFix64(amount))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, fusdAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), fusdSigner},
		shouldRevert,
	)

}

func loadFUSD(ftAddress string) []byte {
	return []byte(test.ReplaceImports(
		string(test.ReadFile(fusdContractPath)),
		map[string]*regexp.Regexp{
			ftAddress: test.FungibleTokenAddressPlaceholder,
		},
	))
}

func replaceAddressPlaceholders(code, ftAddress, fusdAddress string) []byte {
	return []byte(test.ReplaceImports(
		code,
		map[string]*regexp.Regexp{
			ftAddress:  test.FungibleTokenAddressPlaceholder,
			fusdAddress: test.FUSDAddressPlaceHolder,
		},
	))
}

func SetupAccountScript(ftAddress, fusdAddress string) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(fusdSetupAccountPath)),
		ftAddress,
		fusdAddress,
	)
}

func MintFUSDTransaction(fungibleTokenAddress, fusdAddress flow.Address) []byte {
	return replaceAddressPlaceholders(
		string(test.ReadFile(fusdMintTokensPath)),
		fungibleTokenAddress.String(),
		fusdAddress.String(),
	)
}
