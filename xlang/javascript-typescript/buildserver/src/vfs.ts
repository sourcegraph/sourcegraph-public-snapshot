import { FileSystem, FileInfo } from 'javascript-typescript-langserver/src/fs';
import * as filepath from 'path';
import * as fs from 'fs';

/**
 * LayeredFileSystem is a layered file system that builds a composite
 * virtual FS made from the ordered layering of its constituent file
 * systems. File systems earlier in the list take precendence over
 * ones later in the list.
 */
export class LayeredFileSystem implements FileSystem {
	filesystems: FileSystem[];

	constructor(filesystems: FileSystem[]) {
		this.filesystems = filesystems;
	}

	readDir(path: string, callback: (err: Error | null, result?: FileInfo[]) => void) {
		this._readDir(path).then((result) => {
			callback(null, result);
		}, callback);
	}

	readFile(path: string, callback: (err: Error | null, result?: string) => void) {
		this._readFile(path).then((result) => {
			callback(null, result);
		}, callback);
	}

	private async _readDir(path: string): Promise<FileInfo[]> {
		const finfo: FileInfo[] = [];
		const foundNames: { [key: string]: boolean } = {};
		let oneSuccess = false;
		for (let i = 0; i < this.filesystems.length; i++) {
			const f = this.filesystems[i];
			try {
				const newinfos = await readDir(f, path);
				oneSuccess = true;
				for (const newinfo of newinfos) {
					if (!foundNames[newinfo.name]) {
						finfo.push(newinfo);
						foundNames[newinfo.name] = true;
					}
				}
			} catch (e) {
				if (i === this.filesystems.length - 1 && !oneSuccess) {
					throw e;
				}
			}
		}
		return finfo;
	}

	private async _readFile(path: string): Promise<string> {
		for (let i = 0; i < this.filesystems.length; i++) {
			const f = this.filesystems[i];
			try {
				return await readFile(f, path);
			} catch (e) {
				if (i === this.filesystems.length - 1) {
					throw e;
				}
			}
		}
		throw new Error("readFile: no filesystems present");
	}
}

/**
 * LocalRootedFileSystem is a FileSystem implementation backed by the
 * local file system. It differs from the LocalFileSystem class in the
 * language server repository in that it exposes a chrooted filesystem
 * (but don't rely on it to enforce security).
 */
export class LocalRootedFileSystem implements FileSystem {

	private root: string;

	constructor(root: string) {
		this.root = root
	}

	readDir(path: string, callback: (err: Error | null, result?: FileInfo[]) => void) {
		path = filepath.join(this.root, path);
		fs.readdir(path, (err: Error, files: string[]) => {
			if (err) {
				return callback(err)
			}

			Promise.all(
				files.map((f) => {
					return new Promise<FileInfo>((resolve, reject) => {
						fs.stat(filepath.join(path, f), (err: NodeJS.ErrnoException, stats: fs.Stats) => {
							if (err) {
								return reject(err);
							}
							return resolve({
								name: f,
								size: stats.size,
								dir: stats.isDirectory()
							});
						});
					})
				})
			).then((fileInfos) => callback(null, fileInfos), callback);
		});
	}

	readFile(path: string, callback: (err: Error | null, result?: string) => void) {
		path = filepath.join(this.root, path);
		fs.readFile(path, (err: Error, buf: Buffer) => {
			if (err) {
				return callback(err)
			}
			return callback(null, buf.toString()); // assumes UTF-8 encoding
		});
	}

}

/**
 * readFile wraps a call to FileSystem.readFile in a Promise.
 */
export async function readFile(fs: FileSystem, path: string): Promise<string> {
	return new Promise<string>((resolve, reject) => fs.readFile(path, (err, result) => err ? reject(err) : resolve(result)));
}

/**
 * readDir wraps a call to FileSystem.readDir in a Promise.
 */
export async function readDir(fs: FileSystem, path: string): Promise<FileInfo[]> {
	return new Promise<FileInfo[]>((resolve, reject) => fs.readDir(path, (err, result) => err ? reject(err) : resolve(result)));
}

/**
 * walkDirs walks the file tree rooted at root in a FileSystem instance (fs), calling the visit function on each directory.
 */
export async function walkDirs(fs: FileSystem, root: string, visit: (path: string, dirEntries: FileInfo[]) => Promise<void>): Promise<void> {
	const dirEntries = await readDir(fs, root);
	await visit(root, dirEntries);
	return Promise.all(
		dirEntries.map((entry) => entry.dir ? walkDirs(fs, filepath.join(root, entry.name), visit) : Promise.resolve())
	).then(() => { return });
}
