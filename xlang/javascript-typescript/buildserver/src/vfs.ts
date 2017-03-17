import { FileSystem, LocalFileSystem } from 'javascript-typescript-langserver/lib/fs';
import * as path from 'path';
import * as url from 'url';

// TODO Instead of calling all layers, the class could just check the path for node_modules
//      and choose whether to ask the client or the file system (like in the PHP LS)
/**
 * LayeredFileSystem is a layered file system that builds a composite
 * virtual FS made from the ordered layering of its constituent file
 * systems. File systems earlier in the list take precendence over
 * ones later in the list.
 */
export class LayeredFileSystem implements FileSystem {

	constructor(public filesystems: FileSystem[]) {
		if (filesystems.length === 0) {
			throw new Error('Must at least pass one filesystem');
		}
	}

	async getWorkspaceFiles(base?: string): Promise<string[]> {
		// Try all file systems and return the results, if all error reject
		const errors: any[] = [];
		let files: string[] = [];
		// TODO: do in parallel?
		for (const filesystem of this.filesystems) {
			try {
				files = files.concat(await filesystem.getWorkspaceFiles(base));
			} catch (e) {
				errors.push(e)
			}
		}
		if (errors.length === this.filesystems.length) {
			throw Object.assign(new Error('All layered file systems errored: ' + errors.map(e => e.message)), { errors });
		}
		return files;
	}

	async getTextDocumentContent(uri: string): Promise<string> {
		const errors: any[] = [];
		// TODO: do in parallel?
		for (const filesystem of this.filesystems) {
			try {
				return await filesystem.getTextDocumentContent(uri);
			} catch (e) {
				errors.push(e);
			}
		}
		throw Object.assign(new Error('All layered file systems errored: ' + errors.map(e => e.message)), { errors });
	}
}

/**
 * LocalRootedFileSystem is a FileSystem implementation backed by the
 * local file system. It differs from the LocalFileSystem class in the
 * language server repository in that it exposes a mounted filesystem
 * (but don't rely on it to enforce security).
 */
export class LocalRootedFileSystem extends LocalFileSystem {

	constructor(rootPath: string, private mountPath: string) {
		super(rootPath);
	}

	async getWorkspaceFiles(base?: string): Promise<string[]> {
		return (await super.getWorkspaceFiles(base)).map(uri => {
			// Strip mountPath prefix
			const parts = url.parse(uri);
			parts.pathname = path.relative(this.mountPath, parts.pathname);
			return url.format(parts);
		});
	}

	protected resolveUriToPath(uri: string): string {
		// Prefix with mountPath
		return path.join(this.mountPath, super.resolveUriToPath(uri));
	}
}
