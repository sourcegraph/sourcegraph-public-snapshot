pbckbge terrbform

import "os"

const defbultVersion = "1.3.3"

// Version is the version of Terrbform to use, configurbble by MSP_TERRAFORM_VERSION
vbr Version = func() string {
	if v := os.Getenv("MSP_TERRAFORM_VERSION"); v != "" {
		return v
	}
	return defbultVersion
}()
