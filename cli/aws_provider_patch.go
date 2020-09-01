package cli

import (
	"fmt"
	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/mattn/go-zglob"
	"github.com/zclconf/go-cty/cty"
)

func applyAwsProviderPatch(terragruntOptions *options.TerragruntOptions) error {
	if len(terragruntOptions.AwsProviderPatchOverrides) == 0 {
		return errors.WithStackTrace(MissingOverrides(OPT_TERRAGRUNT_OVERRIDE_ATTR))
	}

	terraformFilesInModules, err := findAllTerraformFilesInModules(terragruntOptions)
	if err != nil {
		return err
	}

	for _, terraformFile := range terraformFilesInModules {
		originalTerraformFileContents, err := util.ReadFileAsString(terraformFile)
		if err != nil {
			return err
		}

		updatedTerraformFileContents, err := patchAwsProviderInTerraformCode(originalTerraformFileContents, terraformFile, terragruntOptions.AwsProviderPatchOverrides)
		if err != nil {
			return err
		}

		if originalTerraformFileContents != updatedTerraformFileContents {
			terragruntOptions.Logger.Printf("Patching AWS provider in %s", terraformFile)
			if err := util.WriteFileWithSamePermissions(terraformFile, terraformFile, []byte(updatedTerraformFileContents)); err != nil {
				return err
			}
		}
	}

	return nil
}

// findAllTerraformFiles returns all Terraform source files within the modules being used by this Terragrunt
// configuration. To be more specific, it only returns the source files downloaded for module "xxx" { ... } blocks into
// the .terraform/modules folder; it does NOT return Terraform files for the top-level (AKA "root") module.
//
// NOTE: this method only supports *.tf files right now. Terraform code defined in *.json files is not currently
// supported.d
func findAllTerraformFilesInModules(terragruntOptions *options.TerragruntOptions) ([]string, error) {
	modulesPath := util.JoinPath(terragruntOptions.DataDir(), "modules")

	if !util.FileExists(modulesPath) {
		return nil, nil
	}

	// Ideally, we'd use a builin Go library like filepath.Glob here, but per https://github.com/golang/go/issues/11862,
	// the current go implementation doesn't support treating ** as zero or more directories, just zero or one.
	// So we use a third-party library.
	matches, err := zglob.Glob(fmt.Sprintf("%s/**/*.tf", modulesPath))
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return matches, nil
}

// patchAwsProviderInTerraformCode looks for provider "aws" { ... } blocks in the given Terraform code and overwrites
// the attributes in those provider blocks with the given attributes.
//
// For example, if you passed in the following Terraform code:
//
// provider "aws" {
//    region = var.aws_region
// }
//
// And you set attributesToOverride to map[string]string{"region": "us-east-1"}, then this method will return:
//
// provider "aws" {
//    region = "us-east-1"
// }
//
// This is a temporary workaround for a Terraform bug (https://github.com/hashicorp/terraform/issues/13018) where
// any dynamic values in nested provider blocks are not handled correctly when you call 'terraform import', so by
// temporarily hard-coding them, we can allow 'import' to work.
func patchAwsProviderInTerraformCode(terraformCode string, terraformFilePath string, attributesToOverride map[string]string) (string, error) {
	hclFile, err := hclwrite.ParseConfig([]byte(terraformCode), terraformFilePath, hcl.InitialPos)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	for _, block := range hclFile.Body().Blocks() {
		tokens := block.BuildTokens(nil)
		if len(tokens) > 4 {
			// The tokens should be:
			//   - provider
			//   - "
			//   - aws
			//   - "
			maybeProvider := tokens[0]
			maybeAws := tokens[2]

			if maybeProvider.Type == hclsyntax.TokenIdent && string(maybeProvider.Bytes) == "provider" &&
				maybeAws.Type == hclsyntax.TokenQuotedLit && string(maybeAws.Bytes) == "aws" {

				for key, value := range attributesToOverride {
					block.Body().SetAttributeValue(key, cty.StringVal(value))
				}
			}
		}

	}

	return string(hclFile.Bytes()), nil
}

// Custom error types

type MissingOverrides string

func (err MissingOverrides) Error() string {
	return fmt.Sprintf("You must specify at least one provider attribute to override via the --%s option.", string(err))
}
