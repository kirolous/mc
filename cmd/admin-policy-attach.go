// Copyright (c) 2015-2022 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/minio/madmin-go/v2"
	"github.com/kirolous/mc/pkg/probe"
	"github.com/minio/pkg/console"
)

var adminAttachPolicyFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "user, u",
		Usage: "attach policy to user",
	},
	cli.StringFlag{
		Name:  "group, g",
		Usage: "attach policy to group",
	},
}

var adminPolicyAttachCmd = cli.Command{
	Name:         "attach",
	Usage:        "attach an IAM policy to a user or group",
	Action:       mainAdminPolicyAttach,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Flags:        append(adminAttachPolicyFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS] TARGET POLICY [POLICY...] [--user USER | --group GROUP]

  Exactly one of --user or --group is required.

POLICY:
  Name of the policy on the MinIO server.

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Attach the "readonly" policy to user "james".
     {{.Prompt}} {{.HelpName}} myminio readonly --user james
  2. Attach the "audit-policy" and "acct-policy" policies to group "legal".
     {{.Prompt}} {{.HelpName}} myminio audit-policy acct-policy --group legal
`,
}

// mainAdminPolicyAttach is the handler for "mc admin policy attach" command.
func mainAdminPolicyAttach(ctx *cli.Context) error {
	return userAttachOrDetachPolicy(ctx, true)
}

func userAttachOrDetachPolicy(ctx *cli.Context, attach bool) error {
	if len(ctx.Args()) < 2 {
		showCommandHelpAndExit(ctx, 1) // last argument is exit code
	}

	console.SetColor("PolicyMessage", color.New(color.FgGreen))
	console.SetColor("Policy", color.New(color.FgBlue))

	// Get the alias parameter from cli
	args := ctx.Args()
	aliasedURL := args.Get(0)

	// Put args in PolicyAssociationReq, client checks for validity
	req := madmin.PolicyAssociationReq{
		User:     ctx.String("user"),
		Group:    ctx.String("group"),
		Policies: args.Tail(),
	}

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	fatalIf(err, "Unable to initialize admin connection.")

	var e error
	if attach {
		e = client.AttachPolicy(globalContext, req)
	} else {
		e = client.DetachPolicy(globalContext, req)
	}

	isGroup := false
	if req.User == "" {
		isGroup = true
	}

	userOrGroup := req.User
	if isGroup {
		userOrGroup = req.Group
	}

	if e == nil {
		for _, policy := range req.Policies {
			printMsg(userPolicyMessage{
				op:          ctx.Command.Name,
				Policy:      policy,
				UserOrGroup: userOrGroup,
				IsGroup:     isGroup,
			})
		}
	} else {
		if attach {
			fatalIf(probe.NewError(e), "Unable to attach the policy")
		} else {
			fatalIf(probe.NewError(e), "Unable to detach the policy")
		}
	}
	return nil
}
