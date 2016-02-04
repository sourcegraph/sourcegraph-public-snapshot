package sgx

// This import is imported in a file with a build tag, which means
// `godep save` does not see it. Putting it here, too, lets `godep
// save` see it and vendor it.

import _ "src.sourcegraph.com/sourcegraph/pkg/vcs/ssh"
