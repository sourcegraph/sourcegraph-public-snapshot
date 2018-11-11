// This file is necessary because we want the root tsconfig.json to use {"module": "esnext"}, which means a lone
// gulpfile.ts will fail to compile.
process.env.TS_NODE_COMPILER_OPTIONS = '{"module":"commonjs"}'
require('ts-node/register')
module.exports = require('./gulpfile.ts')
