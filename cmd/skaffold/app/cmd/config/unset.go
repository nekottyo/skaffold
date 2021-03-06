/*
Copyright 2019 The Skaffold Authors

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

package config

import (
	"fmt"
	"io"

	"github.com/GoogleContainerTools/skaffold/cmd/skaffold/app/cmd/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewCmdUnset(out io.Writer) *cobra.Command {
	return commands.
		New(out).
		WithDescription("unset", "Unset a value in the global Skaffold config").
		WithFlags(func(f *pflag.FlagSet) {
			AddConfigFlags(f)
			AddSetFlags(f)
		}).
		ExactArgs(1, doUnset)
}

func doUnset(out io.Writer, args []string) error {
	resolveKubectlContext()
	if err := unsetConfigValue(args[0]); err != nil {
		return err
	}

	logUnsetConfigForUser(out, args[0])
	return nil
}

func logUnsetConfigForUser(out io.Writer, key string) {
	if global {
		fmt.Fprintf(out, "unset global value %s\n", key)
	} else {
		fmt.Fprintf(out, "unset value %s for context %s\n", key, kubecontext)
	}
}

func unsetConfigValue(name string) error {
	return setConfigValue(name, "")
}
