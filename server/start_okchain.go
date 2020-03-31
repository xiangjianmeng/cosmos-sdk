package server

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/node"
)

// okchain full-node start flags
const (
	FlagListenAddr         = "rest.laddr"
	FlagExternalListenAddr = "rest.external_laddr"
	FlagCORS               = "cors"
	FlagMaxOpenConnections = "max-open"
	FlagBackup             = "backup"
	FlagRecover            = "recover"
	FlagHookstartInProcess = "startInProcess"

	// plugin flags
	FlagBackendEnableBackend       = "backend.enable_backend"
	FlagBackendEnableMktCompute    = "backend.enable_mkt_compute"
	FlagBackendLogSql              = "backend.log_sql"
	FlagBackendCleanUpsTime        = "backend.clean_ups_time"
	FlagBacekendOrmEngineType      = "backend.orm_engine.engine_type"
	FlagBackendOrmEngineConnectStr = "backend.orm_engine.connect_str"
)

var (
	backendConf = config.DefaultConfig().BackendConfig
)

//module hook

type fnHookstartInProcess func(ctx *Context) error

type serverHookTable struct {
	hookTable map[string]interface{}
}

var gSrvHookTable = serverHookTable{make(map[string]interface{}, 0)}

func InstallHookEx(flag string, hooker fnHookstartInProcess) {
	gSrvHookTable.hookTable[flag] = hooker
}

//call hooker function
func callHooker(flag string, args ...interface{}) error {
	params := make([]interface{}, 0)
	switch flag {
	case FlagHookstartInProcess:
		{
			//none hook func, return nil
			function, ok := gSrvHookTable.hookTable[FlagHookstartInProcess]
			if !ok {
				return nil
			}
			for _, argv := range args {
				params = append(params, argv)
			}

			if len(params) != 1 {
				return errors.New("too many or less parameter called, want 1")
			}

			//param type check
			p1, ok := params[0].(*Context)
			if !ok {
				return errors.New("wrong param 1 type. want *Context, got" + reflect.TypeOf(params[0]).String())
			}

			//get hook function and call it
			caller := function.(fnHookstartInProcess)
			return caller(p1)
		}
		break
	default:
		break
	}
	return nil
}

//end of hook

func setPID(ctx *Context) {
	pid := os.Getpid()
	f, err := os.OpenFile(filepath.Join(ctx.Config.RootDir, "config", "pid"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		tmos.Exit(err.Error())
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	writer.WriteString(strconv.Itoa(pid))
	writer.Flush()
}

// StopCmd stop the node gracefully
// Tendermint.
func StopCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the node gracefully",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(filepath.Join(ctx.Config.RootDir, "config", "pid"))
			if err != nil {
				errStr := fmt.Sprintf("%s Please finish the process of okchaind through kill -2 pid to stop gracefully", err.Error())
				tmos.Exit(errStr)
			}
			defer f.Close()
			in := bufio.NewScanner(f)
			in.Scan()
			pid, err := strconv.Atoi(in.Text())
			if err != nil {
				errStr := fmt.Sprintf("%s Please finish the process of okchaind through kill -2 pid to stop gracefully", err.Error())
				tmos.Exit(errStr)
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				tmos.Exit(err.Error())
			}
			err = process.Signal(os.Interrupt)
			if err != nil {
				tmos.Exit(err.Error())
			}
			fmt.Println("pid", pid, "has been sent SIGINT")
			return nil
		},
	}
	return cmd
}

func RemoveOldStores(home string, logger log.Logger) error {
	return removeOldStores(home, logger)
}

func removeOldStores(home string, logger log.Logger) error {
	applicationDB := filepath.Join(home, "data/application.db")
	err := os.RemoveAll(applicationDB)
	if err != nil {
		return err
	}
	logger.Info("application db removed...")

	stateDB := filepath.Join(home, "data/state.db")
	err = os.RemoveAll(stateDB)
	if err != nil {
		return err
	}
	logger.Info("state db removed...")

	csDB := filepath.Join(home, "data/cs.wal")
	err = os.RemoveAll(csDB)
	if err != nil {
		return err
	}
	logger.Info("cs wal removed...")

	evidenceDB := filepath.Join(home, "data/evidence.db")
	err = os.RemoveAll(evidenceDB)
	if err != nil {
		return err
	}
	logger.Info("evidence db removed...")

	indexDB := filepath.Join(home, "data/tx_index.db")
	err = os.RemoveAll(indexDB)
	if err != nil {
		return err
	}
	logger.Info("index db removed...")

	stateJson := filepath.Join(home, "data/priv_validator_state.json")
	err = os.Remove(stateJson)
	if err != nil {
		return err
	}
	logger.Info("state json removed...")

	blockDB := filepath.Join(home, "data/blockstore.db")
	err = os.RemoveAll(blockDB)
	if err != nil {
		return err
	}
	logger.Info("blockstore db removed...")

	content := `{
  "height": "0",
  "round": "0",
  "step": 0
}`
	tmos.MustWriteFile(stateJson, []byte(content), 0644)

	return nil
}

var sem *nodeSemaphore

type nodeSemaphore struct {
	done chan struct{}
}

func Stop() {
	sem.done <- struct{}{}
}

// registerRestServerFlags registers the flags required for rest server
func registerRestServerFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().String(FlagListenAddr, "tcp://0.0.0.0:26659", "The address for the rest-server to listen on. (0.0.0.0:0 means any interface, any port)")
	cmd.Flags().String(FlagCORS, "", "Set the rest-server domains that can make CORS requests (* for all)")
	cmd.Flags().Int(FlagMaxOpenConnections, 1000, "The number of maximum open connections of rest-server")
	cmd.Flags().String(FlagExternalListenAddr, "127.0.0.1:26659", "Set the rest-server external ip and port, when it is launched by Docker")
	return cmd
}

// registerOkchainPluginFlags registers the flags required for rest server
func registerOkchainPluginFlags(cmd *cobra.Command) *cobra.Command {

	cmd.Flags().Bool(FlagBackendEnableBackend, backendConf.EnableBackend, "Enable the node's backend plugin")
	cmd.Flags().Bool(FlagBackendEnableMktCompute, backendConf.EnableMktCompute, "Enable kline and ticker calculating")
	cmd.Flags().Bool(FlagBackendLogSql, backendConf.LogSQL, "Enable backend plugin logging sql feature")
	cmd.Flags().String(FlagBackendCleanUpsTime, backendConf.CleanUpsTime, "Backend plugin`s time of cleaning up kline data")
	cmd.Flags().String(FlagBacekendOrmEngineType, backendConf.OrmEngine.EngineType, "Backend plugin`s db (mysql or sqlite3)")
	cmd.Flags().String(FlagBackendOrmEngineConnectStr, backendConf.OrmEngine.ConnectStr, "Backend plugin`s db connect address")

	cmd.Flags().String(node.FlagRollback, "", fmt.Sprintf("Rollback from designated block height --%s=height", node.FlagRollback))
	cmd.Flags().Int64(FlagBackup, 100000, "Specify an interval of block height to back state db")
	cmd.Flags().Int64(FlagRecover, 0, "Specify the state db path to recover")

	return cmd
}
