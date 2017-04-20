
import { FileSystemUpdater } from 'javascript-typescript-langserver/lib/fs';
import { LanguageClient } from 'javascript-typescript-langserver/lib/lang-handler';
import { Logger, NoopLogger, PrefixedLogger } from 'javascript-typescript-langserver/lib/logging';
import { InMemoryFileSystem } from 'javascript-typescript-langserver/lib/memfs';
import { ProjectManager } from 'javascript-typescript-langserver/lib/project-manager';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import * as path from 'path';
import * as url from 'url';
import mkdirp = require('mkdirp');
import { Observable } from '@reactivex/rxjs';
import iterate from 'iterare';
import { fromPairs, toPairs } from 'lodash';
import * as fs from 'mz/fs';
import { Span } from 'opentracing';
import * as semver from 'semver';
import * as yarn from './yarn';
const fetchPackageJson: (packageName: string, options?: { version?: string, fullMetadata?: boolean }) => Promise<PackageJson> = require('package-json');

export interface PackageJson {
	name: string;
	version?: string;
	typings?: string;
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
	private scanned = false;

	/**
	 * Set of package.json URIs _defined_ in the workspace.
	 * This does not include package.jsons of dependencies and also not package.jsons that node_modules are vendored for
	 */
	private packages = new Set<string>();

	constructor(
		private tempDir: string,
		private updater: FileSystemUpdater,
		private inMemoryFileSystem: InMemoryFileSystem,
		private projectManager: ProjectManager,
		private client: LanguageClient,
		private logger: Logger = new NoopLogger()
	) { }

	/**
	 * Returns an Iterable for all package.jsons in the workspace
	 */
	packageJsonUris(): IterableIterator<string> {
		return this.packages.values();
	}

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
	 * @param span OpenTracing span for tracing
	 */
	private scan(span = new Span()): void {
		// Find locations of package.json and node_modules folders
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
		this.packages.clear();
		for (const uri of packageJsons) {
			// Filter package.jsons with vendored node_modules
			if (vendoredPackageJsons.has(uri)) {
				continue;
			}
			// Start installation for the top-level package.json in the background
			if (uri === rootPackageJson) {
				this.ensureForFile(uri, span).catch(err => undefined);
			}
			this.packages.add(uri);
		}
		this.scanned = true;
	}

	/**
	 * Ensures all package.json have been detected, loaded and installations kicked off
	 *
	 * @param span OpenTracing span for tracing
	 */
	async ensureScanned(span = new Span()): Promise<void> {
		await this.updater.ensureStructure(span);
		if (!this.scanned) {
			this.scan(span);
		}
	}

	/**
	 * Gets the content of the closest package.json known to to the DependencyManager in the ancestors of a URI
	 * Call `ensureScanned()` before.
	 */
	async getClosestPackageJson(uri: string, span = new Span()): Promise<PackageJson | undefined> {
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		if (!packageJsonUri) {
			return undefined;
		}
		this.updater.ensure(packageJsonUri, span);
		return JSON.parse(await this.inMemoryFileSystem.getContent(packageJsonUri));
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
			if (this.packages.has(packageJsonUri)) {
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
	 * Call `ensureScanned()` before.
	 *
	 * @param uri URI to a file or directory
	 * @param childOf OpenTracing parent span for tracing
	 */
	private installForFile(uri: string, childOf = new Span()): Promise<void> {
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		if (!packageJsonUri) {
			return Promise.resolve();
		}
		const promise = (async () => {
			const span = childOf.tracer().startSpan('Dependency installation', { childOf });
			span.addTags({ uri, packageJsonUri });
			try {
				const parts = url.parse(packageJsonUri);
				const directory: url.Url = { ...parts, pathname: path.posix.dirname(parts.pathname!) };
				const logger = new PrefixedLogger(this.logger, `inst ${parts.pathname}`);

				// Fetch package.json content
				await this.updater.ensure(packageJsonUri, span);
				const packageJsonContent = this.inMemoryFileSystem.getContent(packageJsonUri);
				const packageJson: PackageJson = JSON.parse(packageJsonContent);

				// Cache key for this package
				const cacheKey = `${packageJson.name}@${packageJson.version}`;
				const neededDependenciesCacheKey = `${cacheKey}:needed_dependencies`;
				const yarnLockCacheKey = `${cacheKey}:yarn.lock`;
				span.addTags({ cacheKey });

				// Before writing package.json to disk, filter out all packages we don't need
				// The only packages we need to install are @types/ packages and packages with a typings field
				// Try to get the filtered dependencies from the cache
				let neededDependencies: { [name: string]: string } | null = await this.client.xcacheGet({ key: neededDependenciesCacheKey });
				span.log({ event: `cache ${neededDependencies ? 'hit' : 'miss'}` });
				if (!neededDependencies) {
					// Else figure it out with NPM registry requests
					neededDependencies = fromPairs(await this.filterDependencies(packageJson, span).toArray().toPromise());
					// Then save it to the cache
					this.client.xcacheSet({ key: neededDependenciesCacheKey, value: neededDependencies });
				}

				span.log({ event: 'needed dependencies', needed: neededDependencies });

				// Rewrite package.json dependencies field to only the packages we need
				packageJson.dependencies = neededDependencies;
				packageJson.devDependencies = undefined;
				packageJson.peerDependencies = undefined;
				packageJson.optionalDependencies = undefined;

				// The directory that yarn will be spawned in
				const cwd = path.join(this.tempDir, 'workspace', uri2path(url.format(directory)));
				const globalFolder = path.join(this.tempDir, 'global', uri2path(url.format(directory)));
				const cacheFolder = path.join(this.tempDir, 'cache', uri2path(url.format(directory)));
				// Create temporary directory
				await new Promise((resolve, reject) => mkdirp(cwd, err => err ? reject(err) : resolve()));
				await new Promise((resolve, reject) => mkdirp(globalFolder, err => err ? reject(err) : resolve()));
				await new Promise((resolve, reject) => mkdirp(cacheFolder, err => err ? reject(err) : resolve()));

				// Write package.json into temporary directory
				await fs.writeFile(path.join(cwd, 'package.json'), JSON.stringify(packageJson));

				// If exists, write yarn.lock into temporary directory to cut down resolve time
				const yarnLockUri = url.format({ ...directory, pathname: path.posix.join(directory.pathname, 'yarn.lock') });

				// First check the cache for (filtered) yarn.lock
				let yarnLock: string | null = await this.client.xcacheGet({ key: cacheKey });

				// If cache miss, check if available in the repo
				if (!yarnLock && this.inMemoryFileSystem.has(yarnLockUri)) {
					await this.updater.ensure(yarnLockUri, span);
					yarnLock = this.inMemoryFileSystem.getContent(yarnLockUri);
				}

				// Write yarn.lock from cache or repo to temporary folder
				if (yarnLock) {
					await fs.writeFile(path.join(cwd, 'yarn.lock'), yarnLock);
				}

				// Spawn yarn process
				await new Promise((resolve, reject) => {
					const yarnProcess = yarn.install({ cwd, globalFolder, cacheFolder, logger }, span);
					this.yarnProcesses.add(yarnProcess);
					yarnProcess.once('success', resolve);
					yarnProcess.once('error', reject);
					yarnProcess.once('exit', () => this.yarnProcesses.delete(yarnProcess));
				});

				// Invalidate the yarn.lock in memory
				this.updater.invalidate(yarnLockUri);

				// Always save the new yarn.lock to the cache
				(async () => {
					try {
						await this.updater.ensure(yarnLockUri);
						yarnLock = this.inMemoryFileSystem.getContent(yarnLockUri);
						this.client.xcacheSet({ key: yarnLockCacheKey, value: yarnLock });
					} catch (err) {
						this.logger.error('Failed to update yarn.lock in cache:', err);
					}
				})();

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
		return promise;
	}

	/**
	 * Reads all dependencies from a package.json file and returns those that contain `.d.ts`
	 * definitions. `@types/` packages are always installed, for all other packages an NPM registry
	 * request is done to find out whether the package.json has a `typings` field.
	 *
	 * @param packageJson Parsed content of a package.json
	 * @param childOf Parent Span for tracing
	 * @return Observable that emits pairs of [package name, version] of packages that need to be installed
	 */
	private filterDependencies(packageJson: PackageJson, childOf = new Span()): Observable<[string, string]> {
		const span = childOf.tracer().startSpan('Filter dependencies', { childOf });
		return Observable.of<keyof PackageJson>('dependencies', 'devDependencies', 'optionalDependencies', 'peerDependencies')
			// Get a stream of package name, version pairs
			.mergeMap(key => toPairs(packageJson[key]) as [string, string][])
			// Exclude file: URI
			.filter(([name, version]) => !version.startsWith('file:'))
			// Remove duplicate packages, if people have them for whatever reason
			.distinct(([name, version]) => name)
			// Filter to only include either @types/ packages or packages with a typings field
			.mergeMap(([name, version]) =>
				(name.startsWith('@types/')
					// @types packages are always needed
					? Observable.of(true)
					// Otherwise only packages with a typings field
					// If the version is not a valid semver range (e.g. GitHub URL), use latest version
					: Observable.from(fetchPackageJson(name, { version: semver.validRange(version) || 'latest', fullMetadata: true }))
						.map(packageJson => !!packageJson.typings)
						// Catch errors and always install packages we failed to fetch the package.json for (e.g. git dependency)
						.catch(err => {
							span.log({ 'event': 'error', 'error.object': err, 'message': err.message, 'stack': err.stack });
							this.logger.error(`Failed to fetch package.json of ${name}@${version} for ${packageJson.name}`, err);
							return [true];
						}))
					.do(needed => {
						span.log({ event: needed ? 'needed' : 'not needed', name, version });
					})
					// Emit the name version pair if it is needed
					.filter(needed => needed)
					.mapTo([name, version])
			)
			.catch(err => {
				span.setTag('error', true);
				span.log({ 'event': 'error', 'error.object': err, 'message': err.message, 'stack': err.stack });
				throw err;
			})
			.finally(() => {
				span.finish();
			});
	}

	/**
	 * Ensures dependencies for the given file or subfolder in the workspace have been installed (at least once)
	 *
	 * @param uri URI to a file or directory
	 * @param childOf OpenTracing parent span for tracing
	 */
	async ensureForFile(uri: string, childOf = new Span()): Promise<void> {
		// Ensure all own package.jsons in the workspace are available under this.packages
		await this.ensureScanned(childOf);
		// Find the closest one in parent directories
		const packageJsonUri = this.getClosestPackageJsonUri(uri);
		childOf.addTags({ packageJsonUri });
		if (!packageJsonUri) {
			return;
		}
		await (this.installations.get(packageJsonUri) || this.installForFile(packageJsonUri, childOf));
	}
}
