import { BuildHandler } from './buildhandler';
import * as server from 'javascript-typescript-langserver/src/server';
import * as util from 'javascript-typescript-langserver/src/util';
const program = require('commander');

const defaultLspPort = 2089;
const numCPUs = require('os').cpus().length;
process.on('uncaughtException', (err: any) => {
	console.error(err);
});

program
	.version('0.0.1')
	.option('-s, --strict', 'Strict mode')
	.option('-p, --port [port]', 'LSP port (' + defaultLspPort + ')', parseInt)
	.option('-c, --cluster [num]', 'Number of concurrent cluster workers (defaults to number of CPUs, ' + numCPUs + ')', parseInt)
	.parse(process.argv);

util.setStrict(program.strict);
const lspPort = program.port || defaultLspPort;
const clusterSize = program.cluster || numCPUs;

server.serve(clusterSize, lspPort, program.strict, () => new BuildHandler());
