package terraform

import "os"

const defaultVersion = "1.3.3"

// Version is the version of Terraform to use, configurable by MSP_TERRAFORM_VERSION
var Version = func() string {
	if v := os.Getenv("MSP_TERRAFORM_VERSION"); v != "" {
		return v
	}
	return defaultVersion
}()
