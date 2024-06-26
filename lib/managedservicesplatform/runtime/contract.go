package runtime

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type Contract = contract.Contract

// ServiceContract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type ServiceContract = contract.ServiceContract

// JobContract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type JobContract = contract.JobContract

// Env carries pre-parsed environment variables and variables requested and
// errors encountered.
type Env = contract.Env
