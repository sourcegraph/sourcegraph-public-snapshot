
import { Observable, Subscription } from '@reactivex/rxjs';
import iterate from 'iterare';
import { FileSystem } from 'javascript-typescript-langserver/lib/fs';
import { RemoteLanguageClient } from 'javascript-typescript-langserver/lib/lang-handler';
import { extractNodeModulesPackageName, PackageJson } from 'javascript-typescript-langserver/lib/packages';
import { ProjectConfiguration } from 'javascript-typescript-langserver/lib/project-manager';
import {
	InitializeParams,
	PackageDescriptor,
	SymbolDescriptor,
	SymbolLocationInformation,
	WorkspaceReferenceParams
} from 'javascript-typescript-langserver/lib/request-type';
import { TypeScriptService, TypeScriptServiceOptions } from 'javascript-typescript-langserver/lib/typescript-service';
import { normalizeUri, uri2path } from 'javascript-typescript-langserver/lib/util';
import { castArray, isEmpty, isEqual } from 'lodash';
import { Span } from 'opentracing';
import * as path from 'path';
import callbackRimraf = require('rimraf');
import * as url from 'url';
import {
	Hover,
	Location,
	SymbolInformation,
	TextDocumentPositionParams
} from 'vscode-languageserver';
import { DependencyManager } from './dependencies';
import { DependencyAwareFileSystem, LocalRootedFileSystem } from './vfs';
import hashObject = require('object-hash');
import jsonpatch from 'fast-json-patch';
const urlRelative: (from: string, to: string) => string = require('url-relative');
const rimraf = Observable.bindNodeCallback<string, void>(callbackRimraf);

interface HasUri {
	uri: string;
}

/**
 * Returns true if the passed argument is an object with a `uri` property
 */
function hasUri(candidate: any): candidate is HasUri {
	return typeof candidate === 'object' && candidate !== null && typeof candidate.uri === 'string';
}

export type BuildHandlerFactory = (client: RemoteLanguageClient, options: BuildHandlerOptions) => BuildHandler;

/**
 * Options to pass to the BuildHandler constructor
 */
export interface BuildHandlerOptions extends TypeScriptServiceOptions {
	/**
	 * The temporary directory to use for this specific workspace/connection,
	 * for example `/tmp/tsjs/worker3/92900ce2-0e47-11e7-93ae-92361f002671`
	 *
	 * Gets created with `mkdir -p` on `initialize` and deleted with `rm -rf` on `shutdown`
	 */
	tempDir: string;
}

/**
 * BuildHandler implements the LanguageHandler interface, providing
 * handler methods for LSP operations. It wraps a TypeScriptService
 * instance (which also implements the LanguageHandler
 * interface). Before calling the corresponding method on the
 * TypeScriptService instance, a BuildHandler method will do the
 * appropriate dependency resolution and fetching. It then rewrites
 * file URIs in the response from the TypeScriptService that refer to
 * files that correspond to fetched dependencies.
 */
export class BuildHandler extends TypeScriptService {
	private remoteFileSystem: FileSystem;

	/**
	 * The options that were passed to the constructor
	 */
	protected options: BuildHandlerOptions;

	/**
	 * Handles installation of dependencies and management of package.jsons in the workspace
	 */
	private dependenciesManager: DependencyManager;

	/**
	 * Subscriptions to unsubscribe on shutdown
	 */
	private subscriptions = new Subscription();

	constructor(client: RemoteLanguageClient, options: BuildHandlerOptions) {
		super(client, options);
	}

	/**
	 * The initialize request is sent as the first request from the client to the server. If the
	 * server receives request or notification before the `initialize` request it should act as
	 * follows:
	 *
	 * - for a request the respond should be errored with `code: -32002`. The message can be picked by
	 * the server.
	 * - notifications should be dropped, except for the exit notification. This will allow the exit a
	 * server without an initialize request.
	 *
	 * Until the server has responded to the `initialize` request with an `InitializeResult` the
	 * client must not sent any additional requests or notifications to the server.
	 *
	 * During the `initialize` request the server is allowed to sent the notifications
	 * `window/showMessage`, `window/logMessage` and `telemetry/event` as well as the
	 * `window/showMessageRequest` request to the client.
	 *
	 * @return Observable of JSON Patches that build an `InitializeResult`
	 */
	initialize(params: InitializeParams, span = new Span()): Observable<jsonpatch.Operation> {
		// Workaround for https://github.com/sourcegraph/sourcegraph/issues/4542
		if (params.rootPath && params.rootPath.startsWith('file://')) {
			params.rootPath = uri2path(params.rootPath);
		}
		return super.initialize(params, span)
			.finally(() => {
				this.dependenciesManager = new DependencyManager(
					this.options.tempDir,
					this.updater,
					this.inMemoryFileSystem,
					this.projectManager,
					this.packageManager,
					this.client,
					this.logger
				);
				// Start dependency installation for the root package.json in the background once all files were detected
				this.subscriptions.add(
					this.updater.ensureStructure()
						.concat(Observable.defer(() => {
							if (this.packageManager.rootPackageJsonUri) {
								return this.dependenciesManager.ensureForFile(this.packageManager.rootPackageJsonUri);
							}
							return Observable.empty<never>();
						}))
						.subscribe(undefined, err => {
							this.logger.error('Error installing dependencies in the background', err);
						})
				);
			});
	}

	/**
	 * Sets up the overlayed file system that includes yarn dependencies
	 */
	protected _initializeFileSystems(accessDisk: boolean): void {
		super._initializeFileSystems(accessDisk);
		this.remoteFileSystem = this.fileSystem;
		const dependencyFileSystem = new LocalRootedFileSystem(this.rootUri, path.join(this.options.tempDir, 'workspace'));
		this.fileSystem = new DependencyAwareFileSystem(dependencyFileSystem, this.remoteFileSystem);
	}

	/**
	 * The shutdown request is sent from the client to the server. It asks the server to shut down,
	 * but to not exit (otherwise the response might not be delivered correctly to the client).
	 * There is a separate exit notification that asks the server to exit.
	 *
	 * @return Observable of JSON Patches that build a `null` result
	 */
	shutdown(params = {}, span = new Span()): Observable<jsonpatch.Operation> {
		this.subscriptions.unsubscribe();
		// Make sure yarn processes do not keep running and recreate the temporary directory
		return Observable.from(this.dependenciesManager.killRunningProcesses())
			.mergeMap(() => {
				// Delete workspace-specific temporary folder with dependencies
				this.logger.log(`Cleaning up temporary folder ${this.options.tempDir} on shutdown`);
				return rimraf(this.options.tempDir);
			})
			.mergeMap(() => super.shutdown(params, span));
	}

	/**
	 * Ensures that dependencies have been installed for all package.jsons that have a dependency to
	 * the package described by the passed PackageDescriptor.
	 *
	 * TODO Install just the passed dependency for the package.json
	 */
	private async _ensureDependency(dependency: PackageDescriptor, dependeeName?: string, span = new Span()): Promise<void> {
		await this.updater.ensureStructure(span).toPromise();
		await Promise.all(
			iterate(this.packageManager.packageJsonUris()).map(async uri => {
				try {
					const packageJson: PackageJson | undefined = await this.packageManager.getPackageJson(uri, span).toPromise();
					if (!dependeeName || packageJson.name === dependeeName) {
						await this.dependenciesManager.ensureForFile(uri, span);
					}
				} catch (err) {
					// Continue on error
					this.logger.error('Ensuring dependency', dependency, dependeeName, err);
				}
			})
		);
	}

	/**
	 * Rewrites a given workspace URI to a Sourcegraph `git://repo?rev#path` URI
	 */
	private async _rewriteUri(originalUri: string): Promise<string> {
		const originalParts = url.parse(originalUri);

		if (!originalParts.pathname) {
			return originalUri;
		}

		// Is the file part of a package in node_modules?
		const packageName = extractNodeModulesPackageName(originalUri);
		if (!packageName) {
			return originalUri;
		}

		const encodedPackageName = packageName.split('/').map(encodeURIComponent).join('/');

		const packageNameIndex = originalParts.pathname.lastIndexOf('/node_modules/' + encodedPackageName);
		const packageRootUri = url.format({ ...originalParts, pathname: originalParts.pathname.slice(0, packageNameIndex) + `/node_modules/${encodedPackageName}` });
		const packageJsonUri = url.format({ ...originalParts, pathname: originalParts.pathname.slice(0, packageNameIndex) + `/node_modules/${encodedPackageName}/package.json` });

		// Get package.json of dependency
		try {
			await this.updater.ensure(packageJsonUri).toPromise();
		} catch (err) {
			// Can't rewrite URI if package.json ist not available
			return originalUri;
		}
		const packageJson: PackageJson = JSON.parse(this.inMemoryFileSystem.getContent(packageJsonUri));

		// Can't find out repo if package.json does not have a repository field
		if (!packageJson.repository) {
			return originalUri;
		}

		// Example: git://github.com/user/repo?rev#path
		// TODO add rev. yarn doesn't write gitHead to package.json: https://github.com/yarnpkg/yarn/issues/2978
		const sourcegraphUrl: url.Url = {
			protocol: 'git',
			slashes: true,
			host: 'github.com'
		};

		// Check package.json repository field
		if (!packageJson.repository) {
			return originalUri;
		}
		if (typeof packageJson.repository === 'string' && /^\w+\/\w+$/.test(packageJson.repository)) {
			// Parse GitHub shorthand, e.g. npm/npm
			// Pathname contains the repo slug
			sourcegraphUrl.pathname = '/' + packageJson.repository;
		} else {
			// Parse GitHub URL like https://github.com/npm/npm.git
			let gitUrl: string;
			if (typeof packageJson.repository === 'object' && typeof packageJson.repository.url === 'string') {
				gitUrl = packageJson.repository.url;
			} else if (typeof packageJson.repository === 'string') {
				gitUrl = packageJson.repository;
			} else {
				return originalUri;
			}
			const repositoryParts = url.parse(gitUrl);
			// Non-GitHub repos are not supported
			if (!repositoryParts.hostname || !repositoryParts.hostname.endsWith('github.com') || !repositoryParts.pathname) {
				return originalUri;
			}
			// Pathname contains the repo slug, without .git suffix
			sourcegraphUrl.pathname = repositoryParts.pathname.replace(/.git$/, '');
		}

		// Hash contains the file path
		sourcegraphUrl.hash = urlRelative(packageRootUri, originalUri);

		if (packageName.startsWith('@types/')) {
			// Special case: @types/ packages are in a subfolder of DefinitelyTyped, named types/<package name>
			sourcegraphUrl.hash = packageName.substr('@'.length) + '/' + sourcegraphUrl.hash;
		}

		return url.format(sourcegraphUrl);
	}

	/**
	 * Rewrites URIs found in a result that refer to a dependency to global Sourcegraph git://repo?rev#path URIs.
	 *
	 * TODO not needed anymore with textDocument/xdefinition?
	 */
	private async _rewriteUris(result: any): Promise<void> {
		if (Array.isArray(result)) {
			await Promise.all(result.map(element => this._rewriteUris(element)));
		} else if (typeof result === 'object' && result !== null) {
			if (hasUri(result)) {
				result.uri = await this._rewriteUri(result.uri);
			} else {
				await Promise.all(Object.keys(result).map(key => this._rewriteUris(result[key])));
			}
		}
	}

	protected _getDefinitionLocations(params: TextDocumentPositionParams, span = new Span()): Observable<Location> {
		const uri = normalizeUri(params.textDocument.uri);
		// Don't wait, but kickoff background job
		this.dependenciesManager.ensureForFile(uri, span).catch(err => undefined);
		let found = false;
		// First, attempt to get definition before dependencies fetching is finished.
		return super._getDefinitionLocations(params, span)
			.catch<Location, Location>(err => [])
			// Check if at least one definition, else wait for dependency fetching to finish and then retry.
			.do(() => found = true)
			.concat(Observable.defer(() =>
				found
				? Observable.empty<Location>()
				: Observable.from(this.dependenciesManager.ensureForFile(uri, span))
					.mergeMap(() => super._getDefinitionLocations(params, span))
			))
			.mergeMap(location => Observable.from(this._rewriteUris(location)).mapTo(location));
	}

	protected _getSymbolLocationInformations(params: TextDocumentPositionParams, span = new Span()): Observable<SymbolLocationInformation> {
		const uri = normalizeUri(params.textDocument.uri);
		// First, attempt to get definition before dependencies fetching is finished.
		this.dependenciesManager.ensureForFile(uri, span).catch(err => undefined);
		let found = false;
		let externalResults = 0;
		return super._getSymbolLocationInformations(params, span)
			.catch<SymbolLocationInformation, SymbolLocationInformation>(err => [])
			.do(() => found = true)
			// Check if at least one definition, else wait for dependency fetching to finish and then retry.
			.concat(Observable.defer<SymbolLocationInformation>(() =>
				found
				? Observable.empty<SymbolLocationInformation>()
				: Observable.from(this.dependenciesManager.ensureForFile(uri, span))
					.mergeMap(() => super._getSymbolLocationInformations(params, span))
			))
			// Strip locations in node_modules because those are not availabe in the client
			.map(({ symbol, location }) => {
				if (location && location.uri.includes('/node_modules/')) {
					location = undefined;
				}
				if (symbol) {
					// Remove node_modules part from a module name
					// The SymbolDescriptor will be used in the defining repo, where the symbol file path will never contain node_modules
					// It may contain the package name though if the repo is a monorepo with multiple packages
					const regExp = /[^"]*node_modules\//;
					symbol.name = symbol.name.replace(regExp, '');
					symbol.containerName = symbol.containerName.replace(regExp, '');
					symbol.filePath = symbol.filePath.replace(regExp, '');
				}
				return { symbol, location };
			})
			// Remove duplicates
			// These can happen if a repository defines the same symbol in multiple locations with
			// interface merging, because we remove the location field
			// See https://github.com/sourcegraph/sourcegraph/issues/5365#issuecomment-294431395
			.distinct(symbol => hashObject(symbol, { respectType: false } as any))
			// Limit external results without location to limit the amount of workspace/symbol queries done for global j2d
			// We currently only use one result anyway
			.takeWhile(({ location }) => {
				if (!location) {
					externalResults++;
				}
				return externalResults <= 1;
			});
	}

	/**
	 * Returns an Observable for all symbols in a given config that match a given SymbolDescriptor or text query
	 *
	 * @param config The ProjectConfiguration to search
	 * @param query A text or SymbolDescriptor query
	 * @return Observable of [match score, SymbolInformation]
	 */
	protected _getSymbolsInConfig(config: ProjectConfiguration, query?: string | Partial<SymbolDescriptor>, childOf = new Span()): Observable<[number, SymbolInformation]> {
		const symbols = super._getSymbolsInConfig(config, query, childOf);
		if (!query || typeof query === 'string') {
			return symbols;
		}
		// If a SymbolDescriptor query is passed, reduce the Observable to only the result with the highest score
		// Sourcegraph currently only jumps to one seemingly random result for xrepo j2d: https://github.com/sourcegraph/sourcegraph/issues/5721
		return symbols.reduce(([score, info], [s, i]): [number, SymbolInformation] => s > score ? [s, i] : [score, info]);
	}

	protected _getHover(params: TextDocumentPositionParams, span = new Span()): Observable<Hover> {
		const uri = normalizeUri(params.textDocument.uri);
		this.dependenciesManager.ensureForFile(uri, span); // don't wait, but kickoff background job
		return super._getHover(params, span)
			.catch<Hover, Hover>(err => [{ contents: [] }])
			// Check if proper Hover result, else wait for dependency fetching to finish and then retry.
			.mergeMap(hover =>
				hover && !isEmpty(hover.contents) && !isEqual(castArray(hover.contents), [{language: 'typescript', value: 'any'}])
				? [hover]
				: Observable.from(this.dependenciesManager.ensureForFile(uri, span))
					.mergeMap(() => super._getHover(params, span))
			);
	}

	/**
	 * The workspace references request is sent from the client to the server to locate project-wide
	 * references to a symbol given its description / metadata.
	 */
	workspaceXreferences(params: WorkspaceReferenceParams, span = new Span()): Observable<jsonpatch.Operation> {
		const dependeePackageName = params.hints ? params.hints.dependeePackageName : undefined;
		const packageDescriptor = params.query.package;
		// If PackageDescriptor is given, install that package
		return Observable.from(packageDescriptor ? this._ensureDependency(packageDescriptor, dependeePackageName, span) : [null])
			.mergeMap(() => super.workspaceXreferences(params, span));
	}
}
