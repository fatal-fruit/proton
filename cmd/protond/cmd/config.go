package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"proton/app"
)

func initSDKConfig(config *sdk.Config) {
	// Set prefixes
	accountPubKeyPrefix := app.Bech32AddressPrefix + "pub"
	validatorAddressPrefix := app.Bech32AddressPrefix + "valoper"
	validatorPubKeyPrefix := app.Bech32AddressPrefix + "valoperpub"
	consNodeAddressPrefix := app.Bech32AddressPrefix + "valcons"
	consNodePubKeyPrefix := app.Bech32AddressPrefix + "valconspub"

	// Set and seal config
	config.SetBech32PrefixForAccount(app.Bech32AddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}
