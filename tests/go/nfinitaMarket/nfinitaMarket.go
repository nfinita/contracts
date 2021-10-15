package nfinitaMarket

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	sdktest "github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"

	"github.com/nfinita/first-market/cadence/tests/go/nft"
	"github.com/nfinita/first-market/cadence/tests/go/test"
)

const (
	nfinitaMarketContractPath = "../../contracts/NfinitaMarket.cdc"

	nfinitaMarketTransactionRootPath = "../../transactions/nfinitaMarket"
	nfinitaMarketSetupAccountPath    = nfinitaMarketTransactionRootPath + "/setup_account.cdc"
	nfinitaMarketSellItemPath        = nfinitaMarketTransactionRootPath + "/sell_item_kibble.cdc"
	nfinitaMarketBuyItemPath         = nfinitaMarketTransactionRootPath + "/buy_item_kibble.cdc"
	nfinitaMarketRemoveItemPath      = nfinitaMarketTransactionRootPath + "/remove_item.cdc"
)

func DeployContracts(t *testing.T, b *emulator.Blockchain) test.Contracts {
	accountKeys := sdktest.AccountKeyGenerator()

	fungibleTokenAddress, kibbleAddress, kibbleSigner := kibble.DeployContracts(t, b)
	nonFungibleTokenAddress, kittyItemsAddress, kittyItemsSigner := nft.DeployContracts(t, b)

	// Should be able to deploy a contract as a new account with one key.
	nfinitaMarketAccountKey, nfinitaMarketSigner := accountKeys.NewWithSigner()
	nfinitaMarketCode := loadNfinitaMarket(
		fungibleTokenAddress,
		nonFungibleTokenAddress,
		kibbleAddress,
		kittyItemsAddress,
	)

	nfinitaMarketAddress, err := b.CreateAccount(
		[]*flow.AccountKey{nfinitaMarketAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "NfinitaMarket",
				Source: string(nfinitaMarketCode),
			},
		},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// simplify the workflow by having contract addresses also be our initial test storefronts
	nft.SetupAccount(t, b, kittyItemsAddress, kittyItemsSigner, nonFungibleTokenAddress, kittyItemsAddress)
	SetupAccount(t, b, nfinitaMarketAddress, nfinitaMarketSigner, nfinitaMarketAddress)

	return test.Contracts{
		FungibleTokenAddress:    fungibleTokenAddress,
		KibbleAddress:           kibbleAddress,
		KibbleSigner:            kibbleSigner,
		NonFungibleTokenAddress: nonFungibleTokenAddress,
		KittyItemsAddress:       kittyItemsAddress,
		KittyItemsSigner:        kittyItemsSigner,
		NfinitaMarketAddress:    nfinitaMarketAddress,
		NfinitaMarketSigner:     nfinitaMarketSigner,
	}
}

func SetupAccount(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress flow.Address,
	userSigner crypto.Signer,
	nfinitaMarketAddr flow.Address,
) {
	tx := flow.NewTransaction().
		SetScript(nfinitaMarketGenerateSetupAccountScript(nfinitaMarketAddr.String())).
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

// Create a new account with the Kibble and KittyItems resources set up BUT no NfinitaMarket resource.
func CreatePurchaserAccount(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
) (flow.Address, crypto.Signer) {
	userAddress, userSigner, _ := test.CreateAccount(t, b)
	kibble.SetupAccount(t, b, userAddress, userSigner, contracts.FungibleTokenAddress, contracts.KibbleAddress)
	nft.SetupAccount(t, b, userAddress, userSigner, contracts.NonFungibleTokenAddress, contracts.KittyItemsAddress)
	return userAddress, userSigner
}

func CreateAccount(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
) (flow.Address, crypto.Signer) {
	userAddress, userSigner := CreatePurchaserAccount(t, b, contracts)
	SetupAccount(t, b, userAddress, userSigner, contracts.NfinitaMarketAddress)
	return userAddress, userSigner
}

func ListItem(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
	userAddress flow.Address,
	userSigner crypto.Signer,
	tokenID uint64,
	price string,
	shouldFail bool,
) (saleOfferResourceID uint64) {
	tx := flow.NewTransaction().
		SetScript(nfinitaMarketGenerateSellItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewUInt64(tokenID))
	_ = tx.AddArgument(test.CadenceUFix64(price))

	result := test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)

	saleOfferAvailableEventType := fmt.Sprintf(
		"A.%s.NfinitaMarket.SaleOfferAvailable",
		contracts.NfinitaMarketAddress,
	)

	for _, event := range result.Events {
		if event.Type == saleOfferAvailableEventType {
			return event.Value.Fields[1].ToGoValue().(uint64)
		}
	}

	return 0
}

func PurchaseItem(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
	userAddress flow.Address,
	userSigner crypto.Signer,
	storefrontAddress flow.Address,
	tokenID uint64,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(nfinitaMarketGenerateBuyItemScript(contracts)).
		SetGasLimit(200).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewUInt64(tokenID))
	_ = tx.AddArgument(cadence.NewAddress(storefrontAddress))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func RemoveItem(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
	userAddress flow.Address,
	userSigner crypto.Signer,
	tokenID uint64,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(nfinitaMarketGenerateRemoveItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewUInt64(tokenID))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func replaceAddressPlaceholders(codeBytes []byte, contracts test.Contracts) []byte {
	return []byte(test.ReplaceImports(
		string(codeBytes),
		map[string]*regexp.Regexp{
			contracts.FungibleTokenAddress.String():    test.FungibleTokenAddressPlaceholder,
			contracts.KibbleAddress.String():           test.KibbleAddressPlaceHolder,
			contracts.NonFungibleTokenAddress.String(): test.NonFungibleTokenAddressPlaceholder,
			contracts.KittyItemsAddress.String():       test.KittyItemsAddressPlaceHolder,
			contracts.NfinitaMarketAddress.String():    test.NfinitaMarketPlaceholder,
		},
	))
}

func loadNfinitaMarket(
	fungibleTokenAddress,
	nonFungibleTokenAddress,
	kibbleAddress,
	kittyItemsAddress flow.Address,
) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(nfinitaMarketContractPath),
		test.Contracts{
			FungibleTokenAddress:    fungibleTokenAddress,
			KibbleAddress:           kibbleAddress,
			NonFungibleTokenAddress: nonFungibleTokenAddress,
			KittyItemsAddress:       kittyItemsAddress,
		},
	)
}

func nfinitaMarketGenerateSetupAccountScript(nfinitaMarketAddr string) []byte {
	return []byte(test.ReplaceImports(
		string(test.ReadFile(nfinitaMarketSetupAccountPath)),
		map[string]*regexp.Regexp{
			nfinitaMarketAddr: test.NfinitaMarketPlaceholder,
		},
	))
}

func nfinitaMarketGenerateSellItemScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(nfinitaMarketSellItemPath),
		contracts,
	)
}

func nfinitaMarketGenerateBuyItemScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(nfinitaMarketBuyItemPath),
		contracts,
	)
}

func nfinitaMarketGenerateRemoveItemScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(nfinitaMarketRemoveItemPath),
		contracts,
	)
}
