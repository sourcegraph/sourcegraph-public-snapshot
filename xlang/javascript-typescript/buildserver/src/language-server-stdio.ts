#!/usr/bin/env node

import { newConnection, registerLanguageHandler } from 'javascript-typescript-langserver/lib/connection';
import * as util from 'javascript-typescript-langserver/lib/util';

import { BuildHandler } from './buildhandler';

const packageJson = require('../package.json');
var program = require('commander');

process.on('uncaughtException', (err: string) => {
	console.error(err);
});

program
	.version(packageJson.version)
	.option('-s, --strict', 'enables strict mode')
	.option('-t, --trace', 'print all requests and responses')
	.option('-l, --logfile [file]', 'also log to this file (in addition to stderr)')
	.parse(process.argv);

util.setStrict(program.strict);
const connection = newConnection(process.stdin, process.stdout, { trace: program.trace, logfile: program.logfile });
registerLanguageHandler(connection, program.strict, new BuildHandler());
connection.listen();
