package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/href/ssh-each/ssh"
	"github.com/href/ssh-each/stream"
	"github.com/href/ssh-each/term"
	cli "github.com/jawher/mow.cli"
	"github.com/lithammer/dedent"
)

func getApp() *cli.Cli {
	app := cli.App("ssh-each", strings.Trim(dedent.Dedent(`
		Run SSH commands on multiple servers concurrently.
		Servers can be passed via -s/--servers, or STDIN.

		Output Modes (-m/--mode):
		  host      shows server before each outputted line, default
		  plain     show output as-is
		  check     show server and ✓ on success, x on failure, no output
		  check-yes show server and ✓ on success, nothing otherwise
		  check-no  show server and x on success, nothing otherwise
		  exit      show server and exit code, no output
		  slient    show nothing

		Exit Code:
		  ssh-each will return an exit code of 0, if at least one command
		  completed and all completed commands were successful.

		  This can be overwritten by using --exit-ok.
	`), "\r\n"))

	app.Spec = strings.Join([]string{
		"[-t]",
		"[-s=<comma-separated-servers>]",
		"[-w=<workers>]",
		"[-u=<user>]",
		"[-p=<port>]",
		"[-m=<mode>]",
		"[--exit-ok]",
		"COMMAND",
	}, " ")

	exitOK := false
	builder := ssh.CommandBuilder{}
	servers := app.StringOpt("s servers", "", "Comma separated servers")
	workers := app.IntOpt("w workers", 16, "Concurrent SSH processes")
	port := app.IntOpt("p port", 0, "Default port")
	mode := app.StringOpt("m mode", "host", "Output mode")
	app.BoolOptPtr(&builder.TTY, "t tty", false, "Use pseudo-terminal")
	app.BoolOptPtr(&exitOK, "exit-ok", false, "Ignore server command errors")
	app.StringOptPtr(&builder.ExplicitUser, "u user", "", "Default user")
	app.StringArgPtr(&builder.Command, "COMMAND", "", "Command to execute")

	app.Action = func() {

		// Validate inputs
		if *workers <= 0 {
			fmt.Println("Must use at least one worker")
			os.Exit(1)
		}

		if *port < 0 || 65535 < *port {
			fmt.Println("Invalid default port")
			os.Exit(1)
		}
		builder.ExplicitPort = uint16(*port)

		reportMode, ok := term.ReportModeFromString(*mode)
		if !ok {
			fmt.Println("Unknown report mode: ", *mode)
			os.Exit(1)
		}

		reader := term.CombinedReader(*servers)
		if reader == nil {
			fmt.Println("Neither stdin, nor -s/--servers given")
			os.Exit(1)
		}

		// Abort on interrupt
		ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
		mux := stream.NewMux(ctx, uint(*workers))
		rep := term.NewReport(reportMode)

		// Generate commands from --servers and from STDIN
		go func() {
			linked := builder.FromReader(ctx, reader)
			for link := range linked {
				rep.Associate(link.Server, link.Command)
				mux.Submit(link.Command)
			}

			// Once all commands have seized, stop Mux from accepting more
			// commands (this causes the workers to wind down).
			mux.Shut()
		}()

		// Report results
		for result := range mux.Results() {
			rep.On(result)
		}

		if exitOK || rep.Success() {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	return app
}

func main() {
	app := getApp()
	err := app.Run(os.Args)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
