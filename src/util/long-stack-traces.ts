// Patches the global Promise to enable long Error stack traces
// To be required when running mocha with --require
import Bluebird from 'bluebird'
Bluebird.config({ longStackTraces: true })
global.Promise = Bluebird
