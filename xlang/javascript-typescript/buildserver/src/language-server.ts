import { spawn } from 'child_process';

import { BuildHandler } from './buildhandler';
import * as server from 'javascript-typescript-langserver/lib/server';
import * as util from 'javascript-typescript-langserver/lib/util';
import { newConnection } from 'javascript-typescript-langserver/lib/connection';
const program = require('commander');
const packageJson = require('../package.json');

const defaultLspPort = 2089;
const numCPUs = require('os').cpus().length;
process.on('uncaughtException', (err: any) => {
	console.error(err);
});

program
	.version(packageJson.version)
	.option('-s, --strict', 'enabled strict mode')
	.option('-p, --port [port]', 'specifies LSP port to use (' + defaultLspPort + ')', parseInt)
	.option('-c, --cluster [num]', 'number of concurrent cluster workers (defaults to number of CPUs, ' + numCPUs + ')', parseInt)
	.option('-t, --trace', 'print all requests and responses')
	.option('-l, --logfile [file]', 'also log to this file (in addition to stderr)')
	.parse(process.argv);

util.setStrict(program.strict);
const lspPort = program.port || defaultLspPort;
const clusterSize = program.cluster || numCPUs;

const options: server.ServeOptions = {
	clusterSize: clusterSize,
	lspPort: lspPort,
	strict: program.strict,
	trace: program.trace,
	logfile: program.logfile
};

server.serve(options, () => new BuildHandler(), () => {
	const args = [];
	if (options.strict) {
		args.push('--strict');
	}
	if (options.trace) {
		args.push('--trace');
	}
	const cp = spawn(
		process.execPath,
		[...process.execArgv, __dirname + '/language-server-stdio.js', ...args],
		{ stdio: ['pipe', 'pipe', 'inherit'] }
	)
	cp.on('exit', code => {
		console.error(`${cp.pid} exited with ${code}`);
	});
	console.error(`spawned ${cp.pid}`);
	return newConnection(cp.stdout, cp.stdin, options);
});
