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
	"errors"
	"fmt"
	"strings"

	"github.com/minio/cli"
	"github.com/minio/madmin-go/v2"
	"github.com/kirolous/mc/pkg/probe"
)

var adminIDPSetCmd = cli.Command{
	Name:         "set",
	Usage:        "Create/Update an IDP server configuration",
	Before:       setGlobalsFromContext,
	Action:       mainAdminIDPSet,
	Hidden:       true,
	OnUsageError: onUsageError,
	Flags:        globalFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} TARGET ID_TYPE [CFG_NAME] [CFG_PARAMS...]

  ID_TYPE must be one of 'ldap' or 'openid'.

  **DEPRECATED**: This command will be removed in a future version. Please use
  "mc admin idp ldap|openid" instead.

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Create/Update the default OpenID IDP configuration (CFG_NAME is omitted).
     {{.Prompt}} {{.HelpName}} play/ openid \
          client_id=minio-client-app \
          client_secret=minio-client-app-secret \
          config_url="http://localhost:5556/dex/.well-known/openid-configuration" \
          scopes="openid,groups" \
          redirect_uri="http://127.0.0.1:10000/oauth_callback" \
          role_policy="consoleAdmin"
  2. Create/Update configuration for OpenID IDP configuration named "dex_test".
     {{.Prompt}} {{.HelpName}} play/ openid dex_test \
          client_id=minio-client-app \
          client_secret=minio-client-app-secret \
          config_url="http://localhost:5556/dex/.well-known/openid-configuration" \
          scopes="openid,groups" \
          redirect_uri="http://127.0.0.1:10000/oauth_callback" \
          role_policy="consoleAdmin"
  3. Create/Update the LDAP IDP configuration (CFG_NAME must be empty for LDAP).
     {{.Prompt}} {{.HelpName}} play/ ldap \
          server_addr=ldap.corp.min.io:686 \
          lookup_bind_dn=cn=readonly,ou=service_account,dc=min,dc=io \
          lookup_bind_password=mysecretpassword \
          user_dn_search_base_dn=dc=min,dc=io \
          user_dn_search_filter="(uid=%s)" \
          group_search_base_dn=ou=swengg,dc=min,dc=io \
          group_search_filter="(&(objectclass=groupofnames)(member=%d))"

`,
}

func validateIDType(idpType string) {
	if !madmin.ValidIDPConfigTypes.Contains(idpType) {
		fatalIf(probe.NewError(errors.New("invalid IDP type")),
			fmt.Sprintf("IDP type must be one of %v", madmin.ValidIDPConfigTypes))
	}
}

func mainAdminIDPSet(ctx *cli.Context) error {
	if len(ctx.Args()) < 3 {
		showCommandHelpAndExit(ctx, 1)
	}

	args := ctx.Args()

	aliasedURL := args.Get(0)

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	fatalIf(err, "Unable to initialize admin connection.")

	idpType := args.Get(1)
	validateIDType(idpType)

	var cfgName string
	input := args[2:]
	if !strings.Contains(args.Get(2), "=") {
		cfgName = args.Get(2)
		input = args[3:]
	}

	inputCfg := strings.Join(input, " ")

	restart, e := client.AddOrUpdateIDPConfig(globalContext, idpType, cfgName, inputCfg, false)
	fatalIf(probe.NewError(e), "Unable to set IDP config for '%s' to server", idpType)

	// Print set config result
	printMsg(configSetMessage{
		targetAlias: aliasedURL,
		restart:     restart,
	})

	return nil
}
