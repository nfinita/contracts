package metagoodMarket

import (
	"fmt"
	"github.com/nfinita/first-market/cadence/tests/go/fusd"
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
	metagoodMarketContractPath = "../../contracts/MetagoodMarket.cdc"

	metagoodMarketTransactionRootPath = "../../transactions/metagoodMarket"
	metagoodMarketScriptsRootPath     = "../../scripts/metagoodMarket"

	metagoodMarketSetupAccountPath = metagoodMarketTransactionRootPath + "/setup_account.cdc"
	metagoodMarketSellItemPath     = metagoodMarketTransactionRootPath + "/sell_item_fusd.cdc"
	metagoodMarketBuyItemPath      = metagoodMarketTransactionRootPath + "/buy_item_fusd.cdc"
	metagoodMarketRemoveItemPath   = metagoodMarketTransactionRootPath + "/remove_item.cdc"

	metagoodMarketReadCollectionIdsPath = metagoodMarketScriptsRootPath + "/read_collection_ids.cdc"
)

func DeployContracts(t *testing.T, b *emulator.Blockchain) test.Contracts {
	//fungibleTokenAddress, fusdAddress, fusdSigner := fusd.DeployContracts(t, b)
	fungibleTokenAddress, fusdAddress, nonFungibleTokenAddress, metaBearAddress, fusdSigner, metaBearSigner :=
		nft.DeployContracts(t, b)

	accountKeys := sdktest.AccountKeyGenerator()

	// Should be able to deploy a contract as a new account with one key.
	metagoodMarketAccountKey, metagoodMarketSigner := accountKeys.NewWithSigner()
	metagoodMarketCode := loadMetagoodMarket(
		fungibleTokenAddress,
		nonFungibleTokenAddress,
		fusdAddress,
		metaBearAddress,
	)

	metagoodMarketAddress, err := b.CreateAccount(
		[]*flow.AccountKey{metagoodMarketAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "MetagoodMarket",
				Source: string(metagoodMarketCode),
			},
		},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	contracts := test.Contracts{
		FungibleTokenAddress:    fungibleTokenAddress,
		FUSDAddress:             fusdAddress,
		FUSDSigner:              fusdSigner,
		NonFungibleTokenAddress: nonFungibleTokenAddress,
		MetaBearAddress:         metaBearAddress,
		MetaBearSigner:          metaBearSigner,
		MetagoodMarketAddress:   metagoodMarketAddress,
		MetagoodMarketSigner:    metagoodMarketSigner,
	}

	// simplify the workflow by having contract addresses also be our initial test storefronts
	nft.SetupAccount(t, b, metaBearAddress, metaBearSigner, nonFungibleTokenAddress, metaBearAddress)

	communityAddress, _ := CreateAccount(t, b, contracts)
	creatorAddress, _ := CreateAccount(t, b, contracts)

	nft.SetMetaBearSettings(
		t,
		b,
		contracts.NonFungibleTokenAddress,
		metaBearAddress,
		metaBearSigner,
		communityAddress,
		"0.03",
		creatorAddress,
		"0.08",
		"0.02",
		metagoodMarketAddress,
		"0.05",
		"0.025",
		false,
	)

	SetupAccount(t, b, metagoodMarketAddress, metagoodMarketSigner, metagoodMarketAddress, fungibleTokenAddress, fusdAddress)

	return contracts
}

func SetupAccount(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress flow.Address,
	userSigner crypto.Signer,
	metagoodMarketAddr flow.Address,
	ftAddr flow.Address,
	fusdAddr flow.Address,
) {
	tx := flow.NewTransaction().
		SetScript(metagoodMarketGenerateSetupAccountScript(
			metagoodMarketAddr.String(),
			ftAddr.String(),
			fusdAddr.String(),
		)).
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

// Create a new account with the FUSD and MetaBear resources set up BUT no MetagoodMarket resource.
func CreatePurchaserAccount(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
) (flow.Address, crypto.Signer) {
	userAddress, userSigner, _ := test.CreateAccount(t, b)
	fusd.SetupAccount(t, b, userAddress, userSigner, contracts.FungibleTokenAddress, contracts.FUSDAddress)
	fusd.Mint(
		t, b,
		contracts.FungibleTokenAddress,
		contracts.FUSDAddress,
		contracts.FUSDSigner,
		userAddress,
		"100.0",
		false,
	)
	// nft.SetupAccount(t, b, userAddress, userSigner, contracts.NonFungibleTokenAddress, contracts.MetaBearAddress)
	return userAddress, userSigner
}

func CreateAccount(
	t *testing.T,
	b *emulator.Blockchain,
	contracts test.Contracts,
) (flow.Address, crypto.Signer) {
	userAddress, userSigner := CreatePurchaserAccount(t, b, contracts)
	SetupAccount(
		t,
		b,
		userAddress,
		userSigner,
		contracts.MetagoodMarketAddress,
		contracts.FungibleTokenAddress,
		contracts.FUSDAddress,
	)
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
		SetScript(metagoodMarketGenerateSellItemScript(contracts)).
		SetGasLimit(1000).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewUInt64(tokenID))
	_ = tx.AddArgument(test.CadenceUFix64(price))
	_ = tx.AddArgument(cadence.NewAddress(contracts.MetaBearAddress))

	result := test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)

	saleOfferAvailableEventType := fmt.Sprintf(
		"A.%s.MetagoodMarket.SaleOfferCreated",
		contracts.MetagoodMarketAddress,
	)

	for _, event := range result.Events {
		if event.Type == saleOfferAvailableEventType {
			print("EVENT VALUES ====\n")
			print(event.Value.Fields[0].String())
			print("\n")
			print(event.Value.Fields[1].String())
			print("\n")
			print(event.Value.Fields[2].String())
			print("\nEVENT VALUES END ====\n")
			return event.Value.Fields[0].ToGoValue().(uint64)
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
	itemID uint64,
	marketAddress flow.Address,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(metagoodMarketGenerateBuyItemScript(contracts)).
		SetGasLimit(300).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewAddress(contracts.MetaBearAddress))
	_ = tx.AddArgument(cadence.NewUInt64(itemID))
	_ = tx.AddArgument(cadence.NewAddress(marketAddress))

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
		SetScript(metagoodMarketGenerateRemoveItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	_ = tx.AddArgument(cadence.NewAddress(contracts.MetaBearAddress))
	_ = tx.AddArgument(cadence.NewUInt64(tokenID))

	test.SignAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func replaceAddressPlaceholders(codeBytes []byte, contracts test.Contracts) []byte {
	return []byte(test.ReplaceEnvs(test.ReplaceImports(
		string(codeBytes),
		map[string]*regexp.Regexp{
			contracts.FungibleTokenAddress.String():    test.FungibleTokenAddressPlaceholder,
			contracts.FUSDAddress.String():             test.FUSDAddressPlaceHolder,
			contracts.NonFungibleTokenAddress.String(): test.NonFungibleTokenAddressPlaceholder,
			contracts.MetaBearAddress.String():         test.MetaBearAddressPlaceHolder,
			contracts.MetagoodMarketAddress.String():   test.MetagoodMarketPlaceholder,
		},
	), map[string]*regexp.Regexp{
		"MetaBearCollection004": test.PathNamePlaceholder,
	}))
}

func loadMetagoodMarket(
	fungibleTokenAddress,
	nonFungibleTokenAddress,
	fusdAddress,
	metaBearAddress flow.Address,
) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(metagoodMarketContractPath),
		test.Contracts{
			FungibleTokenAddress:    fungibleTokenAddress,
			FUSDAddress:             fusdAddress,
			NonFungibleTokenAddress: nonFungibleTokenAddress,
			MetaBearAddress:         metaBearAddress,
		},
	)
}

func metagoodMarketGenerateSetupAccountScript(metagoodMarketAddr, ftAddress, fusdAddress string) []byte {
	return []byte(test.ReplaceImports(
		string(test.ReadFile(metagoodMarketSetupAccountPath)),
		map[string]*regexp.Regexp{
			metagoodMarketAddr: test.MetagoodMarketPlaceholder,
			ftAddress:          test.FungibleTokenAddressPlaceholder,
			fusdAddress:        test.FUSDAddressPlaceHolder,
		},
	))
}

func metagoodMarketGenerateSellItemScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(metagoodMarketSellItemPath),
		contracts,
	)
}

func metagoodMarketGenerateBuyItemScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(metagoodMarketBuyItemPath),
		contracts,
	)
}

func metagoodMarketGenerateRemoveItemScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(metagoodMarketRemoveItemPath),
		contracts,
	)
}

func ReadCollectionIdsScript(contracts test.Contracts) []byte {
	return replaceAddressPlaceholders(
		test.ReadFile(metagoodMarketReadCollectionIdsPath),
		contracts,
	)
}
