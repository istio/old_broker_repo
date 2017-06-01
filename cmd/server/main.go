// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"istio.io/broker/cmd/shared"
	"istio.io/broker/pkg/server"
)

type serverArgs struct {
	port    uint16
	apiPort uint16
}

func serverCmd(printf, fatalf shared.FormatFn) *cobra.Command {
	sa := &serverArgs{}
	serverCmd := cobra.Command{
		Use:   "server",
		Short: "Starts Broker as a server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runServer(sa, printf, fatalf)
		},
	}
	serverCmd.PersistentFlags().Uint16VarP(&sa.port, "port", "p", 9091, "TCP port to use for Broker's Open Service Broker (OSB) API")
	serverCmd.PersistentFlags().Uint16Var(&sa.apiPort, "apiPort", 9093, "TCP port to use for Broker's gRPC API")
	return &serverCmd
}

func runServer(sa *serverArgs, printf, fatalf shared.FormatFn) {
	osb, err := server.CreateServer()
	if err != nil {
		fatalf("Failed to create server: %s", err.Error())
	}

	osb.Start(sa.port)
	printf("Server started, listening on port %s", sa.port)
	fmt.Println("CTL-C to break out of broker")
}

func getRootCmd(args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "brokers",
		Short: "The Istio broker provides open service broker functionality to the Istio services and external marketplaces",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("'%s' is an invalid argument", args[0])
			}
			return nil
		},
	}
	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// hack to make flag.Parsed return true such that glog is happy
	// about the flags having been parsed
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	/* #nosec */
	_ = fs.Parse([]string{})
	flag.CommandLine = fs

	rootCmd.AddCommand(serverCmd(shared.Printf, shared.Fatalf))
	rootCmd.AddCommand(shared.VersionCmd())

	return rootCmd
}

func main() {
	rootCmd := getRootCmd(os.Args[1:])

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
