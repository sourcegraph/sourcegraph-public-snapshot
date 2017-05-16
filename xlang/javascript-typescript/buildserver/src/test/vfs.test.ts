import * as chai from 'chai';
import iterate from 'iterare';
import * as fs from 'mz/fs';
import * as os from 'os';
import * as path from 'path';
import * as rimraf from 'rimraf';
import * as uuid from 'uuid';
import { LocalRootedFileSystem } from '../vfs';
const assert = chai.assert;

describe('vfs.ts', () => {
	describe('LocalRootedFileSystem', () => {
		let tempDir: string;
		let fileSystem: LocalRootedFileSystem;

		before(async () => {
			tempDir = path.join(os.tmpdir(), uuid.v1());

			await fs.mkdir(tempDir);
			await fs.mkdir(path.join(tempDir, '@types'));
			await fs.mkdir(path.join(tempDir, '@types', 'diff'));
			await fs.writeFile(path.join(tempDir, '@types', 'diff', 'index.d.ts'), 'baz');
			await fs.mkdir(path.join(tempDir, 'node_modules'));
			await fs.mkdir(path.join(tempDir, 'node_modules', 'some_package'));
			await fs.mkdir(path.join(tempDir, 'node_modules', 'some_package', 'src'));
			await fs.writeFile(path.join(tempDir, 'node_modules', 'some_package', 'src', 'function.ts'), 'foo');
			await fs.writeFile(path.join(tempDir, 'tweedledee'), 'hi');
			await fs.writeFile(path.join(tempDir, 'tweedledum'), 'bye');
			await fs.mkdir(path.join(tempDir, 'foo'));
			await fs.writeFile(path.join(tempDir, 'foo', 'bar.ts'), 'baz');

			fileSystem = new LocalRootedFileSystem('file:///a/project/', tempDir);
		});

		after(done => rimraf(tempDir, done));

		describe('getWorkspaceFiles()', () => {
			it('should return all files in the workspace', async () => {
				const files = iterate(await fileSystem.getWorkspaceFiles()).toArray();
				assert.sameMembers(files, [
					'file:///a/project/tweedledee',
					'file:///a/project/tweedledum',
					'file:///a/project/foo/bar.ts',
					'file:///a/project/%40types/diff/index.d.ts',
					'file:///a/project/node_modules/some_package/src/function.ts'
				]);
			});
			it('should return all files under specific root', async () => {
				const files = iterate(await fileSystem.getWorkspaceFiles('file:///a/project/foo')).toArray();
				assert.sameMembers(files, [
					'file:///a/project/foo/bar.ts'
				]);
			});
		});
		describe('getTextDocumentContent()', () => {
			it('should read files denoted by absolute URI', async () => {
				const content = await fileSystem.getTextDocumentContent('file:///a/project/tweedledee');
				assert.equal(content, 'hi');
			});
		});
	});
});
