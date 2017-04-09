
global.Promise = require('bluebird');
import * as cluster from 'cluster';
import { FileLogger } from 'javascript-typescript-langserver/lib/logging';
import { createClusterLogger, serve, ServeOptions } from 'javascript-typescript-langserver/lib/server';
import { TypeScriptServiceOptions } from 'javascript-typescript-langserver/lib/typescript-service';
import * as util from 'javascript-typescript-langserver/lib/util';
import * as os from 'os';
import * as path from 'path';
import * as uuid from 'uuid';
import { BuildHandler } from './buildhandler';
import rimraf = require('rimraf');
const { Tracer } = require('lightstep-tracer');
const program = require('commander');

const defaultLspPort = 2089;
const numCPUs = require('os').cpus().length;

program
	.version(require('../package.json').version)
	.option('-s, --strict', 'enabled strict mode')
	.option('-p, --port [port]', 'specifies LSP port to use (' + defaultLspPort + ')', parseInt)
	.option('-c, --cluster [num]', 'number of concurrent cluster workers (defaults to number of CPUs, ' + numCPUs + ')', parseInt)
	.option('-t, --trace', 'print all requests and responses')
	.option('-l, --logfile [file]', 'also log to this file (in addition to stderr)')
	.option('--color', 'force colored output in logs')
	.option('--no-color', 'disable colored output in logs')
	.parse(process.argv);

util.setStrict(program.strict);
const lspPort = program.port || defaultLspPort;
const clusterSize = program.cluster || numCPUs;

const logger = createClusterLogger(program.logfile ? new FileLogger(program.logfile) : undefined);

// Log unhandled rejections
process.on('unhandledRejection', (err: any) => {
	logger.error(err);
});

// Create Tracer if LightStep environment variables are set
const tracer = process.env.LIGHTSTEP_ACCESS_TOKEN && new Tracer({
	access_token: process.env.LIGHTSTEP_ACCESS_TOKEN,
	component_name: 'xlang-typescript',
	verbosity: 1
});

const options: ServeOptions & TypeScriptServiceOptions = {
	clusterSize,
	lspPort,
	strict: program.strict,
	logMessages: program.trace,
	logger,
	tracer
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
		logger.log(`Cleaning up crashed worker ${worker.id}'s temporary directory ${workerTempDir}`);
		rimraf(workerTempDir, err => {
			if (err) {
				logger.error(`Error cleaning up worker tempdir ${workerTempDir}`, err);
			}
		});
	});
} else {
	processTempDir = path.join(baseTempDir, 'worker' + cluster.worker.id);
}

serve(options, client => {
	// Use a different temporary directory for each connection/workspace that is a subdirectory of the worker tempdir
	// The BuildHandler will create it on `initialize` and delete it on `shutdown`
	const tempDir = path.join(processTempDir, uuid.v1());
	logger.log(`Using ${tempDir} as temporary directory`);
	return new BuildHandler(client, { ...options, tempDir });
});
