package hclfmt

import (
	"github.com/gruntwork-io/terragrunt/cli/commands/common"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/pkg/cli"
)

const (
	CommandName = "hclfmt"

	FlagNameTerragruntHCLFmt = "terragrunt-hclfmt-file"
)

func NewCommand(globalOpts *options.TerragruntOptions) *cli.Command {
	opts := NewOptions(globalOpts)

	command := &cli.Command{
		Name:  CommandName,
		Usage: "Recursively find hcl files and rewrite them into a canonical format.",
		Action: func(ctx *cli.Context) error {
			if err := common.InitialSetup(ctx, globalOpts); err != nil {
				return err
			}

			return Run(opts)
		},
	}

	command.AddFlags(
		&cli.GenericFlag[string]{
			Name:        FlagNameTerragruntHCLFmt,
			Destination: &opts.HclFile,
			Usage:       "The path to a single hcl file that the hclfmt command should run on.",
		},
	)

	return command
}
