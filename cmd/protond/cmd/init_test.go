package cmd_test

import (
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/stretchr/testify/require"
	"proton/app"
	"proton/cmd/protond/cmd"
	"testing"
)

const (
	flagNodeHome = "--home"
	flagChainId  = "--chain-id"
)

func TestInitCmd(t *testing.T) {
	var (
		val     = "val"
		chainId = "proton"
	)

	steps := map[string][]string{
		"init": {
			"init",
			"test",
			flagChainId,
			"proton",
			flagNodeHome,
			app.DefaultNodeHome,
		},
		"config-chain-id": {
			"config",
			"chain-id",
			chainId,
			flagNodeHome,
			app.DefaultNodeHome,
		},
		"config-keyring": {
			"config",
			"keyring-backend",
			"test",
			flagNodeHome,
			app.DefaultNodeHome,
		},
		"add-key": {
			"keys",
			"add",
			val,
			flagNodeHome,
			app.DefaultNodeHome,
		},
	}

	initCmd, _ := cmd.NewRootCmd()

	initCmd.SetArgs(steps["init"])
	require.NoError(t, svrcmd.Execute(initCmd, "protond", app.DefaultNodeHome))
}
