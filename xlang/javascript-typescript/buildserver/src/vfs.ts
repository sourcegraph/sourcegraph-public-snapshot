import iterate from 'iterare';
import { FileSystem, LocalFileSystem } from 'javascript-typescript-langserver/lib/fs';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import * as path from 'path';

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

	async getWorkspaceFiles(base?: string): Promise<Iterable<string>> {
		// Try all file systems and return the results, if all error reject
		const errors: any[] = [];
		let files = iterate<string>([]);
		// TODO: do in parallel?
		for (const filesystem of this.filesystems) {
			try {
				files = files.concat(await filesystem.getWorkspaceFiles(base));
			} catch (e) {
				errors.push(e);
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
 * A LocalFileSystem that uses a given mountPath as the root
 */
export class LocalRootedFileSystem extends LocalFileSystem {

	private rootPath: string;

	constructor(rootUri: string, private mountPath: string) {
		super(rootUri);
		this.rootPath = uri2path(rootUri);
	}

	protected resolveUriToPath(uri: string): string {
		// Compute the path relative to the root path and mount it on the mount path instead
		const filePath = super.resolveUriToPath(uri);
		const relative = path.relative(this.rootPath, filePath);
		const mounted = path.join(this.mountPath, relative);
		return mounted;
	}
}
