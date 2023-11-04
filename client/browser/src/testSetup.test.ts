/* eslint-disable @typescript-eslint/no-var-requires */
/* eslint-disable @typescript-eslint/no-require-imports */

import '@sourcegraph/testing/src/jestDomMatchers'

// MessageChannel is not defined in the Vitest jsdom environment.
if (!global.MessageChannel) {
    global.MessageChannel = require('worker_threads').MessageChannel
}
