
global.Promise = require('bluebird');
import { BuildHandler } from './buildhandler';
import { serve, ServeOptions } from 'javascript-typescript-langserver/lib/server';
import { TypeScriptServiceOptions } from 'javascript-typescript-langserver/lib/typescript-service';
import * as util from 'javascript-typescript-langserver/lib/util';
import { RemoteLanguageClient } from 'javascript-typescript-langserver/lib/lang-handler';
import * as cluster from 'cluster';
import * as os from 'os';
import * as path from 'path';
import * as uuid from 'uuid';
import rimraf = require('rimraf');
const program = require('commander');
const packageJson = require('../package.json');

const defaultLspPort = 2089;
const numCPUs = require('os').cpus().length;

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

const options: ServeOptions & TypeScriptServiceOptions = {
	clusterSize: clusterSize,
	lspPort: lspPort,
	strict: program.strict,
	trace: program.trace,
	logfile: program.logfile
};

// Every LSP connection gets a temporary directory in the form of /tmp/tsjs/worker#/uuid

/** Base directory for all processes */
const baseTempDir = path.join(os.tmpdir(), 'tsjs');

/** Base directory for the current process (worker or master) */
let processTempDir: string;

if (cluster.isMaster) {
	processTempDir = path.join(baseTempDir, 'master');
	// If a worker crashes, `rm -rf` its whole temporary directory with all workspace-specific subdirectories
	// A new worker will get forked by serve()
	cluster.on('exit', (worker, code, signal) => {
		const workerTempDir = path.join(baseTempDir, 'worker' + worker.id);
		console.error(`Cleaning up crashed worker ${worker.id}'s temporary directory ${workerTempDir}`);
		rimraf(workerTempDir, err => {
			if (err) {
				console.error(`Error cleaning up worker tempdir ${workerTempDir}`, err);
			}
		});
	});
} else {
	processTempDir = path.join(baseTempDir, 'worker' + cluster.worker.id);
}

serve(options, connection => {
	// Use a different temporary directory for each connection/workspace that is a subdirectory of the worker tempdir
	// The BuildHandler will create it on `initialize` and delete it on `shutdown`
	const tempDir = path.join(processTempDir, uuid.v1());
	console.error(`Using ${tempDir} as temporary directory`);
	return new BuildHandler(new RemoteLanguageClient(connection), { ...options, tempDir });
});
