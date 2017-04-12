#!/usr/bin/env node

import { MessageEmitter, MessageLogOptions, MessageWriter, registerLanguageHandler, RegisterLanguageHandlerOptions } from 'javascript-typescript-langserver/lib/connection';
import { RemoteLanguageClient } from 'javascript-typescript-langserver/lib/lang-handler';
import { FileLogger, StderrLogger } from 'javascript-typescript-langserver/lib/logging';
import * as util from 'javascript-typescript-langserver/lib/util';
import * as os from 'os';
import * as path from 'path';
import * as uuid from 'uuid';
import { isNotificationMessage } from 'vscode-jsonrpc/lib/messages';
import { BuildHandler, BuildHandlerOptions } from './buildhandler';
const { Tracer } = require('lightstep-tracer');
const program = require('commander');

// Log unhandled rejections
process.on('unhandledRejection', (err: any) => {
	logger.error(err);
});

program
	.version(require('../package.json').version)
	.option('-s, --strict', 'enables strict mode')
	.option('-t, --trace', 'print all requests and responses')
	.option('-l, --logfile [file]', 'log to this file')
	.option('--color', 'force colored output in logs')
	.option('--no-color', 'disable colored output in logs')
	.parse(process.argv);

util.setStrict(program.strict);

// Create Tracer if LightStep environment variables are set
const tracer = process.env.LIGHTSTEP_ACCESS_TOKEN && new Tracer({
	access_token: process.env.LIGHTSTEP_ACCESS_TOKEN,
	component_name: 'xlang-typescript',
	verbosity: 0
});

const logger = program.logfile ? new FileLogger(program.logfile) : new StderrLogger();

// The LSP connection gets a temporary directory in the form of /tmp/tsjs/stdio/uuid
// The BuildHandler will create it on `initialize` and delete it on `shutdown`
const tempDir = path.join(process.env.CACHE_DIR || path.join(os.tmpdir(), 'tsjs', 'stdio'), uuid.v1());
logger.log(`Using ${tempDir} as temporary directory`);

const options: BuildHandlerOptions & MessageLogOptions & RegisterLanguageHandlerOptions = {
	tempDir,
	tracer,
	logger,
	logMessages: program.trace,
	strict: program.strict
};

const messageEmitter = new MessageEmitter(process.stdin, options);
const messageWriter = new MessageWriter(process.stdout, options);
const remoteClient = new RemoteLanguageClient(messageEmitter, messageWriter);
const handler = new BuildHandler(remoteClient, options);

// Add an exit notification handler to kill the process
messageEmitter.on('message', message => {
	if (isNotificationMessage(message) && message.method === 'exit') {
		logger.log(`Exit notification`);
		process.exit(0);
	}
});

registerLanguageHandler(messageEmitter, messageWriter, handler,	options);
