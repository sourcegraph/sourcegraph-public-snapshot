
import { FileSystemUpdater } from 'javascript-typescript-langserver/lib/fs';
import { Logger, NoopLogger, PrefixedLogger } from 'javascript-typescript-langserver/lib/logging';
import { InMemoryFileSystem } from 'javascript-typescript-langserver/lib/memfs';
import { ProjectManager } from 'javascript-typescript-langserver/lib/project-manager';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import * as path from 'path';
import * as url from 'url';
import mkdirp = require('mkdirp');
import iterate from 'iterare';
import * as fs from 'mz/fs';
import { Span } from 'opentracing';
import * as yarn from './yarn';

export interface PackageJson {
	name: string;
	version?: string;
	repository?: string | { type: string, url: string };
	dependencies?: {
		[packageName: string]: string;
	};
	devDependencies?: {
		[packageName: string]: string;
	};
	peerDependencies?: {
		[packageName: string]: string;
	};
	optionalDependencies?: {
		[packageName: string]: string;
	};
}

/**
 * Matches:
 *
 *     /foo/node_modules/(bar)/index.d.ts
 *     /foo/node_modules/bar/node_modules/(baz)/index.d.ts
 *     /foo/node_modules/(@types/bar)/index.ts
 */
const PACKAGE_NAME_REGEXP = /.*\/node_modules\/((?:@[^\/]+\/)?[^\/]+)\/.*$/;

/**
 * Returns the name of a package that a file is contained in
 */
export function getPackageName(uri: string): string | undefined {
	const match = decodeURIComponent(url.parse(uri).pathname || '').match(PACKAGE_NAME_REGEXP);
	return match && match[1] || undefined;
}

export class DependencyManager {

	/**
	 * Map from package.json URI to a promise that is fulfilled as soon as the installation for that package.json is completed
	 */
	private installations = new Map<string, Promise<void>>();

	/**
	 * Set of running yarn process to kill on dispose
	 */
	private yarnProcesses = new Set<yarn.YarnProcess>();

	/**
	 * Fulfilled when the workspace was scanned for package.json files, they were fetched and parsed and installations kicked off
	 */
	private scanned?: Promise<void>;

	/**
	 * Map from package.json URI to package.json content of packages _defined_ in the workspace.
	 * This does not include package.jsons of dependencies and also not package.jsons that node_modules are vendored for
	 */
	packages = new Map<string, PackageJson>();

	/**
	 * Whether we should refuse a `workspace/symbol` request because we found that we are in DefinitelyTyped
	 */
	puntWorkspaceSymbol = false;

	constructor(
		private tempDir: string,
		private updater: FileSystemUpdater,
		private inMemoryFileSystem: InMemoryFileSystem,
		private projectManager: ProjectManager,
		private logger: Logger = new NoopLogger()
	) { }

	/**
	 * Disposes the DependencyManager and kills all running yarn processes
	 */
	async killRunningProcesses(): Promise<void> {
		this.logger.log(`Killing ${this.yarnProcesses.size} running yarn processes on dispose`);
		await Promise.all(
			iterate(this.yarnProcesses)
				.map(async yarnProcess => new Promise(resolve => {
					yarnProcess.once('exit', resolve);
					yarnProcess.kill('SIGKILL');
				})
		));
	}

	/**
	 * Scans the workspace to find all packages _defined_ in the workspace, saves the content in `packages`
	 * For each found package, installation is started in the background and tracked in `installations`
	 *
	 * @param childOf OpenTracing parent span for tracing
	 */
	private async scan(childOf = new Span()): Promise<void> {
		const promise = (async () => {
			const span = childOf.tracer().startSpan('Scan dependencies', { childOf });
			try {
				// Find locations of package.json and node_modules folders
				await this.updater.ensureStructure(span);
				const vendoredPackageJsons = new Set<string>();
				const packageJsons = new Set<string>();
				let rootPackageJson: string | undefined;
				let rootPackageJsonLevel = Infinity;
				for (const uri of this.inMemoryFileSystem.uris()) {
					const parts = url.parse(uri);
					if (!parts.pathname) {
						continue;
					}
					// Search for package.json files _not_ inside node_modules
					if (parts.pathname.endsWith('/package.json') && !parts.pathname.includes('/node_modules/')) {
						packageJsons.add(uri);
						// If the current root package.json is further nested than this one, replace it
						const level = parts.pathname.split('/').length;
						if (level < rootPackageJsonLevel) {
							rootPackageJson = uri;
							rootPackageJsonLevel = level;
						}
					}
					// Collect vendored node_modules folders found to filter package.jsons
					const nodeModulesIndex = parts.pathname.indexOf('/node_modules/');
					if (nodeModulesIndex !== -1) {
						vendoredPackageJsons.add(url.format({ ...parts, pathname: uri.slice(0, nodeModulesIndex) + '/package.json' }));
					}
				}
				this.logger.log(`Found ${packageJsons.size} package.json in workspace, ${vendoredPackageJsons.size} vendored node_modules`);
				this.logger.log(`Root package.json: ${rootPackageJson}`);
				// Filter package.jsons with vendored node_modules
				await Promise.all(
					iterate(packageJsons)
						.filter(uri => !vendoredPackageJsons.has(uri))
						.map(async uri => {
							// Fetch package.json content
							await this.updater.ensure(uri, span);
							const packageJson = this.inMemoryFileSystem.getContent(uri);
							const parsedPackageJson: PackageJson = JSON.parse(packageJson);
							// Don't do a workspace/symbol search for DefinitelyTyped
							if (parsedPackageJson.name === 'definitely-typed') {
								this.puntWorkspaceSymbol = true;
							}
							this.packages.set(uri, parsedPackageJson);
							// Start installation for the top-level package.json in the background
							if (uri === rootPackageJson) {
								this.ensureForFile(uri, span).catch(err => undefined);
							}
						})
				);
			} catch (err) {
				this.scanned = undefined;
				throw err;
			}
		})();
		this.scanned = promise;
		return promise;
	}

	/**
	 * Ensures all package.json have been detected, loaded and installations kicked off
	 *
	 * @param childOf OpenTracing parent span for tracing
	 */
	async ensureScanned(childOf = new Span()): Promise<void> {
		const span = childOf.tracer().startSpan('Ensure scanned dependencies', { childOf });
		try {
			await (this.scanned || this.scan(span));
		} catch (err) {
			span.setTag('error', true);
			span.log({ 'event': 'error', 'error.object': err, 'message': err.message, 'stack': err.stack });
			throw err;
		} finally {
			span.finish();
		}
	}

	/**
	 * Gets the content of the closest package.json known to to the DependencyManager in the ancestors of a URI
	 */
	getClosestPackageJson(uri: string): PackageJson | undefined {
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		if (!packageJsonUri) {
			return undefined;
		}
		return this.packages.get(packageJsonUri);
	}

	/**
	 * Walks the parent directories of a given URI to find the first package.json that is known to the InMemoryFileSystem
	 *
	 * TODO return multiple nested package.jsons https://github.com/sourcegraph/sourcegraph/issues/5038
	 *
	 * @param uri URI of a file or directory in the workspace
	 * @return The found package.json or undefined if none found
	 */
	getClosestPackageJsonUri(uri: string): string | undefined {
		const parts = url.parse(uri);
		while (true) {
			if (!parts.pathname) {
				return undefined;
			}
			const packageJsonUri = url.format({ ...parts, pathname: path.posix.join(parts.pathname, 'package.json') });
			if (this.inMemoryFileSystem.has(packageJsonUri)) {
				return packageJsonUri;
			}
			if (parts.pathname === '/') {
				return undefined;
			}
			parts.pathname = path.posix.dirname(parts.pathname);
		}
	}

	/**
	 * Installs dependencies for the given file or directory and refetches structure under that directory.
	 *
	 * @param uri URI to a file or directory
	 * @param childOf OpenTracing parent span for tracing
	 */
	private async installForFile(uri: string, childOf = new Span()): Promise<void> {
		await this.updater.ensureStructure();
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		if (!packageJsonUri) {
			return;
		}
		const promise = (async () => {
			const span = childOf.tracer().startSpan('Dependency installation', { childOf });
			span.addTags({ uri, packageJsonUri });
			try {
				const parts = url.parse(packageJsonUri);
				const logger = new PrefixedLogger(this.logger, `inst ${parts.pathname}`);
				const directory: url.Url = { ...parts, pathname: path.posix.dirname(parts.pathname!) };
				// The directory that yarn will be spawned in
				const cwd = path.join(this.tempDir, 'workspace', uri2path(url.format(directory)));
				const globalFolder = path.join(this.tempDir, 'global', uri2path(url.format(directory)));
				const cacheFolder = path.join(this.tempDir, 'cache', uri2path(url.format(directory)));
				// Create temporary directory
				await new Promise((resolve, reject) => mkdirp(cwd, err => err ? reject(err) : resolve()));
				await new Promise((resolve, reject) => mkdirp(globalFolder, err => err ? reject(err) : resolve()));
				await new Promise((resolve, reject) => mkdirp(cacheFolder, err => err ? reject(err) : resolve()));
				// Fetch package.json content
				await this.updater.ensure(packageJsonUri, span);
				// Write package.json into temporary directory
				await fs.writeFile(path.join(cwd, 'package.json'), this.inMemoryFileSystem.getContent(packageJsonUri));
				// Spawn yarn process
				// TODO return Observable instead of converting to Promise
				await new Promise((resolve, reject) => {
					const yarnProcess = yarn.install({ cwd, globalFolder, cacheFolder, logger }, span);
					this.yarnProcesses.add(yarnProcess);
					yarnProcess.once('success', resolve);
					yarnProcess.once('error', reject);
					yarnProcess.once('exit', () => this.yarnProcesses.delete(yarnProcess));
				});
				// Refetch file structure under node_modules directory
				this.updater.invalidateStructure();
				this.updater.fetchStructure().catch(err => undefined);
				// Require a refresh of module structure
				this.projectManager.invalidateModuleStructure();
				this.projectManager.ensureModuleStructure().catch(err => undefined);
			} catch (err) {
				this.installations.delete(packageJsonUri);
				span.setTag('error', true);
				span.log({ 'event': 'error', 'error.object': err, 'message': err.message, 'stack': err.stack });
				throw err;
			} finally {
				span.finish();
			}
		})();
		this.installations.set(packageJsonUri, promise);
		await promise;
	}

	/**
	 * Ensures dependencies for the given file or subfolder in the workspace have been installed (at least once)
	 *
	 * @param uri URI to a file or directory
	 * @param childOf OpenTracing parent span for tracing
	 */
	async ensureForFile(uri: string, childOf = new Span()): Promise<void> {
		const span = childOf.tracer().startSpan('Ensure dependency installation', { childOf });
		span.addTags({ uri });
		try {
			await this.updater.ensureStructure();
			const packageJsonUri = this.getClosestPackageJsonUri(uri);
			span.addTags({ packageJsonUri });
			if (!packageJsonUri) {
				return;
			}
			await (this.installations.get(packageJsonUri) || this.installForFile(packageJsonUri, span));
		} catch (err) {
			span.setTag('error', true);
			span.log({ 'event': 'error', 'error.object': err, 'message': err.message, 'stack': err.stack });
			throw err;
		} finally {
			span.finish();
		}
	}
}
