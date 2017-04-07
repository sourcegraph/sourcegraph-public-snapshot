
import * as os from 'os';
import * as uuid from 'uuid';
import rimraf = require('rimraf');
import mkdirp = require('mkdirp');
import { assert } from 'chai';
import * as fs from 'mz/fs';
import * as path from 'path';
import * as sinon from 'sinon';
import { install } from '../yarn';
const { MODULE_CACHE_DIRECTORY } = require('yarn/lib/constants');

describe('yarn.ts', () => {
	describe('install()', () => {
		const tempDir = path.join(os.tmpdir(), uuid.v1());
		const cwd = path.join(tempDir, 'cwd');
		const globalFolder = path.join(tempDir, 'global');
		const cacheFolder = path.join(tempDir, 'cache');
		afterEach(done => {
			rimraf(tempDir, done);
		});
		beforeEach(async () => {
			await new Promise((resolve, reject) => mkdirp(cwd, err => err ? reject(err) : resolve()));
			await new Promise((resolve, reject) => mkdirp(globalFolder, err => err ? reject(err) : resolve()));
			await new Promise((resolve, reject) => mkdirp(cacheFolder, err => err ? reject(err) : resolve()));
			await fs.writeFile(path.join(cwd, 'package.json'), JSON.stringify({
				name: 'mypkg',
				version: '4.0.2',
				dependencies: {
					'is-thirteen': '*'
				}
			}));
		});
		it('should install dependencies and emit a success event', async () => {
			await new Promise((resolve, reject) => install({ cwd, cacheFolder, globalFolder, verbose: true }).once('error', reject).once('success', resolve));
			assert(await fs.exists(path.join(cwd, 'node_modules', 'is-thirteen')), 'Expected node_modules/is-thirteen to exist');
		});
		it('should emit step events', async () => {
			const listener = sinon.spy();
			await new Promise((resolve, reject) => {
				install({ cwd, cacheFolder, globalFolder, verbose: true })
					.once('error', reject)
					.once('success', resolve)
					.on('step', listener);
			});
			sinon.assert.called(listener);
			sinon.assert.alwaysCalledWith(listener, sinon.match.object);
		});
		it('should use the passed cache instead of the default one', done => {
			install({ cwd, cacheFolder, globalFolder, verbose: true })
				.on('verbose', log => {
					// yarn logs "Copying MODULE_CACHE_DIRECTORY/... to ..." for every file during linking phase
					assert.notInclude(log, MODULE_CACHE_DIRECTORY);
				})
				.once('error', done)
				.once('success', done);
		});
	});
});
