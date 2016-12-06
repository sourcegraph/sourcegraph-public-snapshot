/// <reference path="../node_modules/vscode/thenable.d.ts" />

import * as net from 'net';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import * as cluster from 'cluster';

import { newConnection, registerLanguageHandler } from 'javascript-typescript-langserver/src/connection';
import * as util from 'javascript-typescript-langserver/src/util';

import { BuildHandler } from './buildhandler';

const program = require('commander');

process.on('uncaughtException', (err) => {
	console.error(err);
});

const defaultLspPort = 2089;
const numCPUs = require('os').cpus().length;

program
	.version('0.0.1')
	.option('-s, --strict', 'Strict mode')
	.option('-p, --port [port]', 'LSP port (' + defaultLspPort + ')', parseInt)
	.option('-c, --cluster [num]', 'Number of concurrent cluster workers (defaults to number of CPUs, ' + numCPUs + ')', parseInt)
	.parse(process.argv);

util.setStrict(program.strict);
const lspPort = program.port || defaultLspPort;
const clusterSize = program.cluster || numCPUs;

if (cluster.isMaster) {
	console.error(`Master node process spawning ${clusterSize} workers`)
	for (let i = 0; i < clusterSize; ++i) {
		const worker = cluster.fork().on('disconnect', () => {
			console.error(`worker ${worker.process.pid} disconnect`)
		});
	}

	cluster.on('exit', (worker, code, signal) => {
		const reason = code === null ? signal : code;
		console.error(`worker ${worker.process.pid} exit (${reason})`);
	});
} else {
	console.error('Listening for incoming LSP connections on', lspPort);
	var server = net.createServer((socket) => {
		const connection = newConnection(socket, socket);
		registerLanguageHandler(connection, program.strict, new BuildHandler());
		connection.listen();
	});

	server.listen(lspPort);
}
