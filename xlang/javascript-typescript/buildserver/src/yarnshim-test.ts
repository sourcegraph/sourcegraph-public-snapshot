import { install } from './yarnshim';
import { FileSystem } from 'javascript-typescript-langserver/lib/fs';
import * as path from 'path';
import * as cluster from 'cluster';
import * as process from 'process';
import * as rimraf from 'rimraf';

// clusterSize is the number of parallel processes running yarn.
const clusterSize = 10;

// yarntestdir is the temp directory that contains all files that yarn writes to disk.
const yarntestdir = '/tmp/yarnshim-test'

/**
 * InMemFileSystem mocks the FileSystem interface with in-memory contents.
 */
class InMemFileSystem implements FileSystem {
	private files = new Map<string, string>();

	constructor(root: string) { }

	async getWorkspaceFiles(): Promise<string[]> {
		return Array.from(this.files.keys());
	}

	async getTextDocumentContent(uri: string): Promise<string> {
		if (!this.files.has(uri)) {
			throw new Error(`Unknown file ${uri}`);
		}
		return this.files.get(uri);
	}

	addFile(uri: string, content: string) {
		this.files.set(uri, content);
	}
}

/**
 * rmrf is a Promise wrapper around rimraf.
 */
function rmrf(p: string): Promise<void> {
	return new Promise<void>((resolve, reject) => {
		rimraf(p, (err) => {
			if (err) {
				return reject(err);
			} else {
				return resolve();
			}
		});
	});
}

/**
 * runInstall runs one `yarn install`.
 */
async function runInstall(): Promise<void> {
	const fs = new InMemFileSystem('/');
	fs.addFile('/package.json', '{ "name": "test", "dependencies": {"tslint": "4.1.1"} }');
	await install(fs, '/', path.join(yarntestdir, 'global'), path.join(yarntestdir, 'workspace'));
}

async function main() {
	if (cluster.isMaster) {
		await rmrf(yarntestdir);
		console.error(`Master node process spawning ${clusterSize} workers`)

		let workersFinished = 0;
		let workersFailed = 0;
		for (let i = 0; i < clusterSize; ++i) {
			const worker = cluster.fork().on('disconnect', () => {
				console.error(`worker ${worker.process.pid} disconnect`)
			});
		}

		cluster.on('exit', (worker, code, signal) => {
			const reason = code === null ? signal : code;
			console.error(`worker ${worker.process.pid} exit (${reason})`);

			workersFinished++;
			if (code !== 0) {
				workersFailed++;
			}
			if (workersFinished === clusterSize) {
				if (workersFailed > 0) {
					console.error("\nTest result: FAILED\n");
				} else {
					console.error("\nTest result: PASS\n");
				}
				process.exit(workersFailed);
			}
		});
	} else {
		console.error("running `yarn install`");
		runInstall().then(() => {
			console.error("worker done");
			process.exit();
		}, (err) => {
			console.error("worker error:", err);
			process.exit(1);
		});
	}
}

main();
