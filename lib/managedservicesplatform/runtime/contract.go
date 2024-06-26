package runtime

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type Contract = contract.Contract

type ServiceContract = contract.ServiceContract
type JobContract = contract.JobContract

// Env carries pre-parsed environment variables and variables requested and
// errors encountered.
type Env = contract.Env
