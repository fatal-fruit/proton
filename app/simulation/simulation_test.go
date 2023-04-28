package simulation_test

//
//import (
//	"os"
//	"testing"
//	"time"
//
//	"cosmossdk.io/simapp"
//	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
//	"github.com/cosmos/cosmos-sdk/x/simulation"
//	"github.com/stretchr/testify/require"
//	abci "github.com/cometbft/cometbft/abci/types"
//	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
//	tmtypes "github.com/cometbft/cometbft/types"
//
//	"proton/app"
//)
//
//func init() {
//	simapp.GetSimulatorFlags()
//}
//
//var defaultConsensusParams = &abci.ConsensusParams{
//	Block: &abci.BlockParams{
//		MaxBytes: 200000,
//		MaxGas:   2000000,
//	},
//	Evidence: &tmproto.EvidenceParams{
//		MaxAgeNumBlocks: 302400,
//		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
//		MaxBytes:        10000,
//	},
//	Validator: &tmproto.ValidatorParams{
//		PubKeyTypes: []string{
//			tmtypes.ABCIPubKeyTypeEd25519,
//		},
//	},
//}
//
//func BenchmarkSimulation(b *testing.B) {
//	simapp.FlagEnabledValue = true
//	simapp.FlagCommitValue = true
//
//	config, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
//	require.NoError(b, err, "simulation setup failed")
//
//	b.Cleanup(func() {
//		db.Close()
//		err = os.RemoveAll(dir)
//		require.NoError(b, err)
//	})
//
//	encoding := app.MakeEncodingConfig()
//
//	app := app.New(
//		logger,
//		db,
//		nil,
//		true,
//		map[int64]bool{},
//		app.DefaultNodeHome,
//		0,
//		encoding,
//		simapp.EmptyAppOptions{},
//	)
//
//	// Run randomized simulations
//	_, simParams, simErr := simulation.SimulateFromSeed(
//		b,
//		os.Stdout,
//		app.BaseApp,
//		simapp.AppStateFn(app.AppCodec(), app.SimulationManager()),
//		simulationtypes.RandomAccounts,
//		simapp.SimulationOperations(app, app.AppCodec(), config),
//		app.ModuleAccountAddrs(),
//		config,
//		app.AppCodec(),
//	)
//
//	// export state and simParams before the simulation error is checked
//	err = simapp.CheckExportSimulation(app, config, simParams)
//	require.NoError(b, err)
//	require.NoError(b, simErr)
//
//	if config.Commit {
//		simapp.PrintStats(db)
//	}
//}