//
package dev

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/spf13/cobra"

	"github.com/nanobox-io/nanobox/config"
	engineutil "github.com/nanobox-io/nanobox/util/engine"
	"github.com/nanobox-io/nanobox/util/server"
	mistutil "github.com/nanobox-io/nanobox/util/server/mist"
)

var (

	//
	deployCmd = &cobra.Command{
		Hidden: true,

		Use:   "deploy",
		Short: "Deploys code to the nanobox",
		Long:  ``,

		PreRun:  boot,
		Run:     deploy,
		PostRun: halt,
	}

	//
	install bool // tells nanobox server to install services
)

//
func init() {
	deployCmd.Flags().BoolVarP(&install, "run", "", false, "Creates your app environment w/o webs or workers")
}

// deploy
func deploy(ccmd *cobra.Command, args []string) {

	// PreRun: boot

	fmt.Printf(stylish.Bullet("Deploying codebase..."))

	// stream deploy output
	go mistutil.Stream([]string{"log", "deploy"}, mistutil.PrintLogStream)

	// listen for status updates
	errch := make(chan error)
	go func() {
		errch <- mistutil.Listen([]string{"job", "deploy"}, mistutil.DeployUpdates)
	}()

	v := url.Values{}
	v.Add("reset", strconv.FormatBool(config.Force))
	v.Add("run", strconv.FormatBool(install))

	// remount the engine file at ~/.nanobox/apps/<app>/<engine> so any new scripts
	// will be used during the deploy
	if err := engineutil.RemountLocal(); err != nil {
		config.Debug("No engine mounted (not found locally).")
	}

	// run a deploy
	if err := server.Deploy(v.Encode()); err != nil {
		server.Fatal("[commands/deploy] server.Deploy() failed", err.Error())
	}

	// wait for a status update (blocking)
	err := <-errch

	//
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	// reset "reloaded" to false after a successful deploy so as NOT to deploy on
	// subsequent runnings of "nanobox dev"
	config.VMfile.ReloadedIs(false)

	// PostRun: halt
}