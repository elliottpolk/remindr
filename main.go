package main

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var (
	version  string
	compiled string = fmt.Sprint(time.Now().Unix())

	cfgFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c", "cfg", "confg"},
		Usage:   "optional path to config file",
	})

	userFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "user",
		Aliases: []string{"u"},
		Usage:   "auth username for the email server (will use from-address if no user provided)",
	})

	passwdFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "password",
		Aliases: []string{"passwd", "p"},
		Usage:   "auth password for the email server",
	})

	fromAddrFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "from.address",
		Aliases: []string{"from-addr", "fa"},
		Usage:   "email address for whom it's is being sent by",
	})

	toAddrsFlag = altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
		Name:    "to.addresses",
		Aliases: []string{"to-addrs", "ta"},
		Usage:   "list of email addresses to send message to",
	})

	smtpHostFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "smtp.host",
		Aliases: []string{"sh", "host"},
		Usage:   "host for the email server",
	})

	smtpPortFlag = altsrc.NewIntFlag(&cli.IntFlag{
		Name:    "smtp.port",
		Aliases: []string{"sp", "port"},
		Value:   25,
		Usage:   "port for the email server",
	})

	subjectFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "message.subject",
		Aliases: []string{"s"},
		Usage:   "subject for reminder message to be sent",
	})

	msgFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "message.body",
		Aliases: []string{"m"},
		Usage:   "reminder message to be sent",
	})
)

func main() {
	ct, err := strconv.ParseInt(compiled, 0, 0)
	if err != nil {
		panic(err)
	}

	app := cli.App{
		Name:      "remindr",
		Usage:     "notify people to do their timesheets",
		Copyright: "Copyright © 2020 Elliott Polk",
		Version:   version,
		Compiled:  time.Unix(ct, -1),
		Flags: []cli.Flag{
			cfgFlag,
			userFlag,
			passwdFlag,
			fromAddrFlag,
			toAddrsFlag,
			smtpHostFlag,
			smtpPortFlag,
			subjectFlag,
			msgFlag,
		},
		Before: func(ctx *cli.Context) error {
			if len(ctx.String(cfgFlag.Name)) > 0 {
				return altsrc.InitInputSourceWithContext(ctx.App.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))(ctx)
			}
			return nil
		},
		Action: func(ctx *cli.Context) error {
			// set logging level
			log.SetLevel(log.InfoLevel)

			faddr := ctx.String(fromAddrFlag.Name)
			if len(faddr) < 1 {
				return cli.Exit(errors.New("a valid from address must be provided"), 1)
			}

			user := ctx.String(userFlag.Name)
			if len(user) < 1 {
				user = faddr
			}

			taddrs := ctx.StringSlice(toAddrsFlag.Name)
			if len(taddrs) < 1 {
				return cli.Exit(errors.New("at least 1 to address must be provided"), 1)
			}

			host := ctx.String(smtpHostFlag.Name)
			if len(host) < 1 {
				return cli.Exit(errors.New("a valid SMTP host must be provided"), 1)
			}

			port := ctx.Int(smtpPortFlag.Name)

			subject := "REMINDER"
			if s := ctx.String(subjectFlag.Name); len(s) > 0 {
				subject = s
			}

			msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\nThis is an automated reminder notification", subject))
			if m := ctx.String(msgFlag.Name); len(m) > 0 {
				msg = []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, m))
			}

			var (
				pwd            = ctx.String(passwdFlag.Name)
				auth smtp.Auth = nil
			)

			if len(pwd) > 0 {
				auth = smtp.PlainAuth("", user, pwd, host)
			}

			if err := smtp.SendMail(fmt.Sprintf("%s:%d", host, port), auth, faddr, taddrs, msg); err != nil {
				return cli.Exit(err, 1)
			}

			log.Info("email sent...")
			return nil
		},
	}

	app.Run(os.Args)

}
