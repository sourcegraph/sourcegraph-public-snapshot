// The configuration and the import of core-js needs to be split into separate files,
// because ESM imports are hoisted to the top before application code.
// It is important though that we configure core-js before it is imported.

import './configure-core-js'
import './polyfill'
