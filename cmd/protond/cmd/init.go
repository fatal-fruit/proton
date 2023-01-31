package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"proton/app"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
)

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			switch {
			case chainID != "":
			case clientCtx.ChainID != "":
				chainID = clientCtx.ChainID
			default:
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			// Get bip39 mnemonic
			var mnemonic string
			recover, _ := cmd.Flags().GetBool(FlagRecover)
			if recover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

			if !overwrite && tmos.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			appGenState := mbm.DefaultGenesis(cdc)

			// Override default denom and set to uproton
			if err = configureStaking(appGenState, clientCtx, app.BondDenom); err != nil {
				return err
			}
			if err = configureGov(appGenState, clientCtx, app.BondDenom); err != nil {
				return err
			}
			if err = configureMint(appGenState, clientCtx, app.BondDenom); err != nil {
				return err
			}

			appState, err := json.MarshalIndent(appGenState, "", " ")
			if err != nil {
				return errors.Wrap(err, "Failed to marshal default genesis state")
			}

			genDoc := &types.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = types.GenesisDocFromFile(genFile)
				if err != nil {
					return errors.Wrap(err, "Failed to read genesis doc from file")
				}
			}

			genDoc.ChainID = chainID
			genDoc.Validators = nil
			genDoc.AppState = appState

			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return errors.Wrap(err, "Failed to export genesis file")
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))

	return err
}

func configureStaking(appGenState map[string]json.RawMessage, ctx client.Context, denom string) error {
	var stakingGenesis stakingtypes.GenesisState

	stakingRaw := appGenState[stakingtypes.ModuleName]
	err := ctx.Codec.UnmarshalJSON(stakingRaw, &stakingGenesis)
	if err != nil {
		return err
	}
	stakingGenesis.Params.BondDenom = denom
	modifiedStakingStr, err := ctx.Codec.MarshalJSON(&stakingGenesis)
	if err != nil {
		return err
	}

	appGenState[stakingtypes.ModuleName] = modifiedStakingStr
	return nil
}

func configureGov(appGenState map[string]json.RawMessage, ctx client.Context, denom string) error {
	var govGenesis govtypesv1.GenesisState

	govRaw := appGenState[govtypes.ModuleName]
	err := ctx.Codec.UnmarshalJSON(govRaw, &govGenesis)
	if err != nil {
		return err
	}
	govGenesis.Params.MinDeposit = append(govGenesis.DepositParams.MinDeposit, sdk.NewCoin(denom, sdk.NewInt(10000000)))
	modifiedGovStr, err := ctx.Codec.MarshalJSON(&govGenesis)
	if err != nil {
		return err
	}

	appGenState[govtypes.ModuleName] = modifiedGovStr
	return nil
}

func configureMint(appGenState map[string]json.RawMessage, ctx client.Context, denom string) error {
	var mintGenesis minttypes.GenesisState

	mintRaw := appGenState[minttypes.ModuleName]
	err := ctx.Codec.UnmarshalJSON(mintRaw, &mintGenesis)
	if err != nil {
		return err
	}
	mintGenesis.Params.MintDenom = denom
	modifiedMintStr, err := ctx.Codec.MarshalJSON(&mintGenesis)
	if err != nil {
		return err
	}

	appGenState[minttypes.ModuleName] = modifiedMintStr
	return nil
}
