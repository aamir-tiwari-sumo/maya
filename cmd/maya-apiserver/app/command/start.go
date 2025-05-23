/*
Copyright 2017 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"sync"

	"strings"
	"syscall"
	"time"

	"github.com/aamir-tiwari-sumo/maya/cmd/maya-apiserver/app/config"
	"github.com/aamir-tiwari-sumo/maya/cmd/maya-apiserver/app/server"
	"github.com/aamir-tiwari-sumo/maya/cmd/maya-apiserver/cstor-operator/spc"
	bd "github.com/aamir-tiwari-sumo/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/aamir-tiwari-sumo/maya/pkg/blockdeviceclaim/v1alpha1"
	env "github.com/aamir-tiwari-sumo/maya/pkg/env/v1alpha1"
	install "github.com/aamir-tiwari-sumo/maya/pkg/install/v1alpha1"
	"github.com/aamir-tiwari-sumo/maya/pkg/usage"
	"github.com/aamir-tiwari-sumo/maya/pkg/util"
	"github.com/aamir-tiwari-sumo/maya/pkg/version"
	errors "github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var (
	helpText = `
Usage: m-apiserver start [options]

  Starts maya-apiserver and runs until an interrupt is received.

  The maya apiserver's configuration primarily comes from the config
  files used, but a subset of the options may also be passed directly
  as CLI arguments, listed below.

General Options :

  -bind=<addr>
    The address the server will bind to for all of its various network
    services. The individual services that run bind to individual
    ports on this address. Defaults to the loopback 127.0.0.1.

  -config=<path>
    The path to either a single config file or a directory of config
    files to use for configuring maya api server. This option may be
    specified multiple times. If multiple config files are used, the
    values from each will be merged together. During merging, values
    from files found later in the list are merged over values from
    previously parsed files.

  -log-level=<level>
    Specify the verbosity level of maya api server's logs. Valid values include
    DEBUG, INFO, and WARN, in decreasing order of verbosity. The
    default is INFO.
 `
)

// gracefulTimeout controls how long we wait before forcefully terminating
const gracefulTimeout = 5 * time.Second

// CmdStartOptions is a cli implementation that runs a maya apiserver.
// The command will not end unless a shutdown message is sent on the
// ShutdownCh. If two messages are sent on the ShutdownCh it will forcibly
// exit.
type CmdStartOptions struct {
	BindAddr   string
	LogLevel   string
	ConfigPath string
	ShutdownCh <-chan struct{}
	args       []string

	// TODO
	// Check if both maya & httpServer instances are required ?
	// Can httpServer or maya embed one of the other ?
	// Need to take care of shuting down & graceful exit scenarios !!
	maya       *server.MayaApiServer
	httpServer *server.HTTPServer
}

// NewCmdStart creates start command for maya-apiserver
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start maya-apiserver",
		Long:  helpText,

		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd, &options), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.BindAddr, "bind", "", options.BindAddr,
		"IP Address to bind for maya apiserver.")

	cmd.Flags().StringVarP(&options.LogLevel, "log-level", "", options.LogLevel,
		"Log level for maya apiserver DEBUG INFO WARN.")

	cmd.Flags().StringVarP(&options.ConfigPath, "config", "", options.ConfigPath,
		"Path to a single config file or directory.")

	return cmd
}

// This fn performs preflight checks related to m-api server
// so that server can start properly
func performPreflightChecks() error {
	return checkForNDMrelatedCRDs()
}

// checks existence of NDM related CRDs
func checkForNDMrelatedCRDs() error {
	_, err := bdc.NewKubeClient().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Errorf("precheck for bdc failed: %v", err)
	}
	_, err = bd.NewKubeClient().List(metav1.ListOptions{})
	if err != nil {
		return errors.Errorf("precheck for bd failed: %v", err)
	}
	return nil
}

// Run does tasks related to mayaserver.
func Run(cmd *cobra.Command, c *CmdStartOptions) error {
	klog.Infof("Initializing maya-apiserver...")

	var ControllerMutex = sync.RWMutex{}
	// Read and merge with default configuration
	mconfig := c.readMayaConfig()
	if mconfig == nil {
		return errors.New("Unable to load the configuration")
	}

	//TODO Setup Log Level

	// perform preflight checks
	err := performPreflightChecks()
	if err != nil {
		return errors.Errorf("preflight checks failed: %v", err)
	}

	// Setup Maya server
	if err = c.setupMayaServer(mconfig); err != nil {
		return err
	}
	defer c.maya.Shutdown()

	// Check and shut down at the end
	defer func() {
		if c.httpServer != nil {
			c.httpServer.Shutdown()
		}
	}()

	// Compile Maya server information for output later
	info := make(map[string]string)
	info["version"] = fmt.Sprintf("%s%s", mconfig.Version, mconfig.VersionPrerelease)
	info["log level"] = mconfig.LogLevel
	info["region"] = fmt.Sprintf("%s (DC: %s)", mconfig.Region, mconfig.Datacenter)

	// Sort the keys for output
	infoKeys := make([]string, 0, len(info))
	for key := range info {
		infoKeys = append(infoKeys, key)
	}
	sort.Strings(infoKeys)

	// Maya server configuration output
	padding := 18
	klog.Info("Maya api server configuration:\n")
	for _, k := range infoKeys {
		klog.Infof(
			"%s%s: %s",
			strings.Repeat(" ", padding-len(k)),
			strings.Title(k),
			info[k])
	}
	klog.Infof("")

	// Output the header that the server has started
	klog.Info("Maya api server started! Log data will stream in below:\n")

	// start storage pool controller
	go func() {
		err := spc.Start(&ControllerMutex)
		if err != nil {
			klog.Errorf("Failed to start storage pool controller: %s", err.Error())
		}

	}()

	if env.Truthy(env.OpenEBSEnableAnalytics) {
		usage.New().Build().InstallBuilder(true).Send()
		go usage.PingCheck()
	}

	// Wait for exit
	if c.handleSignals(mconfig) > 0 {
		return errors.New("Ungraceful exit")
	}

	return nil
}

func (c *CmdStartOptions) readMayaConfig() *config.MayaConfig {
	// Load the configuration
	mconfig := config.DefaultMayaConfig()

	if c.ConfigPath != "" {
		current, err := config.LoadMayaConfig(c.ConfigPath)
		if err != nil {
			klog.Errorf(
				"Error loading configuration from %s: %s", c.ConfigPath, err)
			return nil
		}

		// The user asked us to load some config here but we didn't find any,
		// so we'll complain but continue.
		if current == nil || reflect.DeepEqual(current, &config.MayaConfig{}) {
			klog.Warningf("No configuration loaded from %s", c.ConfigPath)
		}

		if mconfig == nil {
			mconfig = current
		} else {
			mconfig = mconfig.Merge(current)
		}
	}

	// Merge any CLI options over config file options

	// Set the version info
	mconfig.Revision = version.GetGitCommit()
	mconfig.Version = version.GetVersion()
	mconfig.VersionPrerelease = version.GetBuildMeta()

	// Set the details from command line
	if c.BindAddr != "" {
		mconfig.BindAddr = c.BindAddr
	}
	if c.LogLevel != "" {
		mconfig.LogLevel = c.LogLevel
	}

	// Normalize binds, ports, addresses, and advertise
	if err := mconfig.NormalizeAddrs(); err != nil {
		klog.Errorf(err.Error())
		return nil
	}

	return mconfig
}

// setupMayaServer is used to start Maya server
func (c *CmdStartOptions) setupMayaServer(mconfig *config.MayaConfig) error {
	klog.Info("Starting maya api server ...")

	// run maya installer
	installErrs := install.SimpleInstaller().Install()
	if len(installErrs) != 0 {
		klog.Errorf("failed to apply resources: %+v", installErrs)
		return errors.New("failed to apply resources")
	}

	klog.Info("resources applied successfully by installer")

	// clean older CASTemplates and RunTasks
	cleanErr := install.SimpleInstaller().Clean()
	if cleanErr != nil {
		klog.Errorf("failed to clean old resources: %s", cleanErr.Error())
	}

	// Setup maya service i.e. maya api server
	maya, err := server.NewMayaApiServer(mconfig, os.Stdout)
	if err != nil {
		klog.Errorf("failed to start api server: %+v", err)
		return err
	}

	c.maya = maya

	// Setup the HTTP server
	http, err := server.NewHTTPServer(maya, mconfig, os.Stdout)
	if err != nil {
		maya.Shutdown()
		klog.Errorf("failed to start http server: %+v", err)
		return err
	}

	c.httpServer = http
	return nil
}

// handleSignals blocks until we get an exit-causing signal
func (c *CmdStartOptions) handleSignals(mconfig *config.MayaConfig) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)

	// Wait for a signal
WAIT:
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-c.ShutdownCh:
		sig = os.Interrupt
	}
	klog.Infof("Caught signal: %v", sig)

	// Skip any SIGPIPE signal (See issue #1798)
	if sig == syscall.SIGPIPE {
		goto WAIT
	}

	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		if conf := c.handleReload(mconfig); conf != nil {
			// Update the value only, not address
			*mconfig = *conf
		}
		goto WAIT
	}

	// Check if we should do a graceful leave
	graceful := false
	if sig == os.Interrupt && mconfig.LeaveOnInt {
		graceful = true
	} else if sig == syscall.SIGTERM && mconfig.LeaveOnTerm {
		graceful = true
	}

	// Bail fast if not doing a graceful leave
	if !graceful {
		return 1
	}

	// Attempt a graceful leave
	gracefulCh := make(chan struct{})
	klog.Info("Gracefully shutting maya api server...")
	go func() {
		if err := c.maya.Leave(); err != nil {
			klog.Errorf("Error: %s", err)
			return
		}
		close(gracefulCh)
	}()

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// handleReload is invoked when we should reload our configs, e.g. SIGHUP
// TODO
// The current reload code is very basic.
// Add ways to reload the orchestrator & plugins without shuting down the
// process
func (c *CmdStartOptions) handleReload(mconfig *config.MayaConfig) *config.MayaConfig {

	klog.Info("Reloading maya api server configuration...")

	newConf := c.readMayaConfig()
	if newConf == nil {
		klog.Error("Failed to reload config")
		return mconfig
	}

	//TODO Change the log level dynamically
	klog.Infof("Log level is : %s", strings.ToUpper(newConf.LogLevel))

	return newConf
}
