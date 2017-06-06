import { Observable } from '@reactivex/rxjs';
import { FileSystem, LocalFileSystem } from 'javascript-typescript-langserver/lib/fs';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import { noop } from 'lodash';
import { Span } from 'opentracing';
import * as path from 'path';

/**
 * Will use a local file system for dependencies (URIs that contain `node_modules`)
 * and a remote file system otherwise
 */
export class DependencyAwareFileSystem implements FileSystem {

	constructor(private dependencyFs: FileSystem, private workspaceFs: FileSystem) {}

	/**
	 * Gets aggregated workspace files from both file systems.
	 * Errors only if both file systems error.
	 */
	getWorkspaceFiles(base?: string, span = new Span()): Observable<string> {
		const errors: any[] = [];
		return Observable.of(this.dependencyFs, this.workspaceFs)
			.mergeMap(filesystem =>
				filesystem.getWorkspaceFiles(base, span)
					.catch(err => {
						errors.push(err);
						return [];
					})
			)
			.do(noop, noop, () => {
				if (errors.length === 2) {
					throw Object.assign(new Error('Failed to get workspace files: ' + errors.map(e => e.message).join(', ')), { errors });
				}
			});
	}

	/**
	 * Gets the content from the dependency file system if the URI includes `node_modules` or `yarn.lock`, from remote file system otherwise.
	 * Also falls back to remote file system of dependency file system errors.
	 */
	getTextDocumentContent(uri: string, span = new Span()): Observable<string> {
		if (uri.includes('/node_modules/') || uri.endsWith('/yarn.lock')) {
			return this.dependencyFs.getTextDocumentContent(uri, span)
				// If the dependency file system fails, fallback to the remote file system
				// node_modules is sometimes vendored
				.catch(err => this.workspaceFs.getTextDocumentContent(uri, span));
		}
		return this.workspaceFs.getTextDocumentContent(uri, span);
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
