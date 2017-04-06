
import { FileSystemUpdater } from 'javascript-typescript-langserver/lib/fs';
import { Logger, NoopLogger, PrefixedLogger } from 'javascript-typescript-langserver/lib/logging';
import { InMemoryFileSystem, ProjectManager } from 'javascript-typescript-langserver/lib/project-manager';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import * as path from 'path';
import * as url from 'url';
import mkdirp = require('mkdirp');
import { spawn } from 'child_process';
import iterate from 'iterare';
import * as fs from 'mz/fs';
import { Span } from 'opentracing';

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
								this.ensureForFile(uri).catch(err => undefined);
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
			span.log({ 'event': 'error', 'error.object': err });
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
	 * Walks the parent directories of a given URI to find the first package.json that is know to the DependencyManager
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
			const packageJson = url.format({ ...parts, pathname: path.posix.join(parts.pathname, 'package.json') });
			if (this.packages.has(packageJson)) {
				return packageJson;
			}
			if (parts.pathname === '/') {
				return undefined;
			}
			parts.pathname = path.posix.dirname(parts.pathname);
		}
	}

	/**
	 * Installs dependencies for the given file or directory and refetches structure under that directory.
	 * Does not depend on a call to `ensureScanned()`
	 *
	 * @param uri URI to a file or directory
	 * @param childOf OpenTracing parent span for tracing
	 */
	private async installForFile(uri: string, childOf = new Span()): Promise<void> {
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		if (!packageJsonUri) {
			return Promise.resolve();
		}
		const promise = (async () => {
			const span = childOf.tracer().startSpan('Dependency installation', { childOf });
			span.addTags({ uri, packageJsonUri });
			try {
				const logger = new PrefixedLogger(this.logger, `Dependency installation ${packageJsonUri}`);
				const parts = url.parse(packageJsonUri);
				const directory: url.Url = { ...parts, pathname: path.posix.dirname(parts.pathname!) };
				// The directory that yarn will be spawned in
				const cwd = path.join(this.tempDir, 'workspace', uri2path(url.format(directory)));
				const globalDir = path.join(this.tempDir, 'global', uri2path(url.format(directory)));
				const cacheDir = path.join(this.tempDir, 'cache', uri2path(url.format(directory)));
				// Create temporary directory
				await new Promise((resolve, reject) => mkdirp(cwd, err => err ? reject(err) : resolve()));
				await new Promise((resolve, reject) => mkdirp(globalDir, err => err ? reject(err) : resolve()));
				await new Promise((resolve, reject) => mkdirp(cacheDir, err => err ? reject(err) : resolve()));
				// Fetch package.json content
				await this.updater.ensure(packageJsonUri, span);
				// Write package.json into temporary directory
				await fs.writeFile(path.join(cwd, 'package.json'), this.inMemoryFileSystem.getContent(packageJsonUri));
				await new Promise((resolve, reject) => {
					// Spawn yarn process
					const yarn = spawn(process.execPath, [
						path.resolve(__dirname, '..', 'node_modules', 'yarn', 'bin', 'yarn.js'),
						'install',
						'--ignore-scripts',  // Don't run package.json scripts
						'--ignore-platform', // Don't error on failing platform checks
						'--ignore-engines',  // Don't check package.json engines field
						'--no-bin-links',    // Don't create bin symlinks
						'--no-lockfile',     // Don't read or create a lockfile
						'--no-emoji',        // Don't use emojis in output
						'--non-interactive', // Don't ask for any user input
						'--no-progress',     // Don't output a progress bar
						// '--link-duplicates', // Use hardlinks instead of copying, not working reliably because of https://github.com/yarnpkg/yarn/issues/2734

						// Use a separate global and cache folders per package.json
						// that we can clean up afterwards and don't interfere with concurrent installations
						'--global-folder', globalDir,
						'--cache-folder', cacheDir
					], { cwd });
					// Forward all output to logger
					yarn.stdout.on('data', chunk => {
						try {
							logger.log((chunk + '').trim());
						} catch (err) {
							reject(err);
						}
					});
					// Capture STDERR output in case of an error
					let stderr = '';
					yarn.stderr.on('data', chunk => {
						try {
							const str = chunk + '';
							stderr += str;
							if (str.startsWith('warning')) {
								logger.warn(str.trim());
							} else {
								logger.error(str.trim());
							}
						} catch (err) {
							reject(err);
						}
					});
					yarn.on('exit', code => {
						if (code === 0) {
							resolve();
						} else {
							reject(Object.assign(new Error(`yarn install failed with exit code ${code}: ${stderr}`), { stderr }));
						}
					});
				});
				// Refetch file structure under node_modules directory
				this.updater.invalidateStructure();
				this.updater.fetchStructure().catch(err => undefined);
				// Require re-fetching of file imports
				this.projectManager.ensuredFilesForHoverAndDefinition.clear();
				// Require re-fetching of module structure
				this.projectManager.ensuredModuleStructure = undefined;
			} catch (err) {
				this.installations.delete(packageJsonUri);
				span.setTag('error', true);
				span.log({ 'event': 'error', 'error.object': err });
				throw err;
			} finally {
				span.finish();
			}
		})();
		this.installations.set(packageJsonUri, promise);
		return promise;
	}

	/**
	 * Ensures dependencies for the given file or subfolder in the workspace have been installed (at least once)
	 *
	 * @param uri URI to a file or directory
	 * @param childOf OpenTracing parent span for tracing
	 */
	async ensureForFile(uri: string, childOf = new Span()): Promise<void> {
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		if (!packageJsonUri) {
			return;
		}
		const span = childOf.tracer().startSpan('Ensure dependency installation', { childOf });
		span.addTags({ uri, packageJsonUri });
		try {
			await (this.installations.get(packageJsonUri) || this.installForFile(packageJsonUri, span));
		} catch (err) {
			span.setTag('error', true);
			span.log({ 'event': 'error', 'error.object': err });
			throw err;
		} finally {
			span.finish();
		}
	}
}
