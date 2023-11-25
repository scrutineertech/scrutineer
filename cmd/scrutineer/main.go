package main

import (
	"errors"
	"log"
	"os"
	"runtime/debug"
	"scrutineer.tech/scrutineer/internal/cli"
	"scrutineer.tech/scrutineer/internal/util"
	"strconv"
	"time"

	urfave "github.com/urfave/cli/v2"
)

var Version string

func main() {
	buildInfoVersion := ""
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		buildInfoVersion = buildInfo.Main.Version
	}

	cliConfig, err := cli.New(util.OrString(os.Getenv("SCRUTINEER_API_ENDPOINT"), "https://api.scrutineer.tech"), util.OrString(Version, buildInfoVersion, "dev"))
	if err != nil {
		log.Fatalf(err.Error())
	}

	// PGP specific calling logic
	args := os.Args
	if len(args) >= 3 && args[2] == "-bsau" {
		cliConfig.SignMsg(os.Stdin, os.Stderr, os.Stdout)
		return
	}
	if len(args) >= 4 && args[2] == "--verify" {
		signature, err := os.Open(args[3])
		if err != nil {
			log.Fatalf(err.Error())
		}

		cliConfig.Verify(os.Stdin, signature, os.Stdout, os.Stderr)
		return
	}

	// Scrutineer specific calls
	app := &urfave.App{
		Name:    "scrutineer",
		Usage:   "Sign and verify git commits",
		Version: cliConfig.Version,
		Commands: []*urfave.Command{
			{
				Name: "login",
				Action: func(context *urfave.Context) error {
					return cliConfig.GithubAuth()
				},
			},
			{
				Name: "logout",
				Action: func(context *urfave.Context) error {
					return cliConfig.Logout()
				},
			},
			{
				Before: func(context *urfave.Context) error {
					return cliConfig.AuthCacheExists()
				},
				Name: "unregister",
				Action: func(context *urfave.Context) error {
					return cliConfig.DeleteMe()
				},
			},
			{
				Before: func(context *urfave.Context) error {
					return cliConfig.AuthCacheExists()
				},
				Name:  "whoami",
				Usage: "returns your logged in user id",
				Action: func(context *urfave.Context) error {
					return cliConfig.Whoami()
				},
			},
			{
				Before: func(context *urfave.Context) error {
					return cliConfig.AuthCacheExists()
				},
				Name:  "realm",
				Usage: "show realm",
				Action: func(context *urfave.Context) error {
					return cliConfig.GetRealm()
				},
			},
			{
				Before: func(context *urfave.Context) error {
					return cliConfig.AuthCacheExists()
				},
				Name:  "trust",
				Usage: "options to trust someone or something",
				Subcommands: []*urfave.Command{
					{
						Name:  "user",
						Usage: "create user trust, default to 1 year",
						Action: func(cCtx *urfave.Context) error {
							trustee := cCtx.Args().First()
							if trustee == "" {
								urfave.ShowSubcommandHelpAndExit(cCtx, 1)
							}
							startTime := cCtx.Timestamp("start")
							if startTime == nil {
								// default to now
								startTime = &time.Time{}
							}
							endTime := cCtx.Timestamp("end")
							if endTime == nil {
								endTime = &time.Time{}
							}
							return cliConfig.CreateUserUserTrust(trustee, *startTime, *endTime)
						},
						Flags: []urfave.Flag{
							&urfave.TimestampFlag{
								Name:   "start",
								Usage:  "Instead of 'now', start trusting at a specific time (UTC)",
								Layout: "2006-01-02T15:04:05",
							},
							&urfave.TimestampFlag{
								Name:   "end",
								Usage:  "Instead of 'now+365d', end trust at a specific time (UTC)",
								Layout: "2006-01-02T15:04:05",
							},
						},
					},
				},
			},
			{
				Before: func(context *urfave.Context) error {
					return cliConfig.AuthCacheExists()
				},
				Name:  "revoke",
				Usage: "revoke [id]",
				Action: func(cCtx *urfave.Context) error {
					trustIdString := cCtx.Args().First()
					if trustIdString == "" {
						urfave.ShowSubcommandHelpAndExit(cCtx, 1)
					}
					// convert trustId to int
					trustId, err := strconv.Atoi(trustIdString)
					if err != nil {
						return errors.New("revoke trust based on an integer")
					}
					return cliConfig.DeleteUserUserTrust(trustId)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
