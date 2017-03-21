#!/usr/bin/env node

import { newConnection, registerLanguageHandler } from 'javascript-typescript-langserver/lib/connection';
import { RemoteLanguageClient } from 'javascript-typescript-langserver/lib/lang-handler';
import * as util from 'javascript-typescript-langserver/lib/util';
import * as path from 'path';
import * as os from 'os';
import * as uuid from 'uuid';

import { BuildHandler } from './buildhandler';

const packageJson = require('../package.json');
var program = require('commander');

program
	.version(packageJson.version)
	.option('-s, --strict', 'enables strict mode')
	.option('-t, --trace', 'print all requests and responses')
	.option('-l, --logfile [file]', 'also log to this file (in addition to stderr)')
	.parse(process.argv);


// The LSP connection gets a temporary directory in the form of /tmp/tsjs/stdio/uuid
// The BuildHandler will create it on `initialize` and delete it on `shutdown`
const tempDir = path.join(os.tmpdir(), 'tsjs', 'stdio', uuid.v1());
console.error(`Using ${tempDir} as temporary directory`);
util.setStrict(program.strict);
const connection = newConnection(process.stdin, process.stdout, { trace: program.trace, logfile: program.logfile });
registerLanguageHandler(connection, new BuildHandler(new RemoteLanguageClient(connection), { strict: program.strict, tempDir }));
connection.listen();
