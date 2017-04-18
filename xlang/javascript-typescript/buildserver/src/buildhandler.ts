
import { Observable } from '@reactivex/rxjs';
import iterate from 'iterare';
import { isCancelledError } from 'javascript-typescript-langserver/lib/cancellation';
import { FileSystem } from 'javascript-typescript-langserver/lib/fs';
import { RemoteLanguageClient } from 'javascript-typescript-langserver/lib/lang-handler';
import {
	InitializeParams,
	PackageDescriptor,
	ReferenceInformation,
	SymbolLocationInformation,
	WorkspaceReferenceParams,
	WorkspaceSymbolParams
} from 'javascript-typescript-langserver/lib/request-type';
import { TypeScriptService, TypeScriptServiceOptions } from 'javascript-typescript-langserver/lib/typescript-service';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import { isEmpty } from 'lodash';
import { Span } from 'opentracing';
import * as path from 'path';
import * as rimraf from 'rimraf';
import * as url from 'url';
import {
	Hover,
	InitializeResult,
	Location,
	SymbolInformation,
	TextDocumentPositionParams
} from 'vscode-languageserver';
import { DependencyManager, getPackageName, PackageJson } from './dependencies';
import { LayeredFileSystem, LocalRootedFileSystem } from './vfs';
import hashObject = require('object-hash');
const urlRelative: (from: string, to: string) => string = require('url-relative');

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

	constructor(client: RemoteLanguageClient, options: BuildHandlerOptions) {
		super(client, options);
	}

	async initialize(params: InitializeParams, span = new Span()): Promise<InitializeResult> {
		// Workaround for https://github.com/sourcegraph/sourcegraph/issues/4542
		if (params.rootPath && params.rootPath.startsWith('file://')) {
			params.rootPath = uri2path(params.rootPath);
		}
		const result = await super.initialize(params, span);
		this.dependenciesManager = new DependencyManager(this.options.tempDir, this.updater, this.inMemoryFileSystem, this.projectManager, this.client, this.logger);
		// Start installation of dependencies in the background
		this.dependenciesManager.ensureScanned(span).catch(err => {
			if (!isCancelledError(err)) {
				this.logger.error('Dependency initialization failed: ', err);
			}
		});
		return result;
	}

	/**
	 * Sets up the overlayed file system that includes yarn dependencies
	 */
	protected _initializeFileSystems(accessDisk: boolean): void {
		super._initializeFileSystems(accessDisk);
		this.remoteFileSystem = this.fileSystem;
		const overlayFs = new LocalRootedFileSystem(this.root, path.join(this.options.tempDir, 'workspace'));
		this.fileSystem = new LayeredFileSystem([overlayFs, this.remoteFileSystem]);
	}

	async shutdown(params = {}, span = new Span()): Promise<null> {
		// Make sure yarn processes do not keep running and recreate the temporary directory
		await this.dependenciesManager.killRunningProcesses();
		// Delete workspace-specific temporary folder with dependencies
		this.logger.log(`Cleaning up temporary folder ${this.options.tempDir} on shutdown`);
		await new Promise((resolve, reject) => rimraf(this.options.tempDir, err => err ? reject(err) : resolve()));
		return await super.shutdown(params, span);
	}

	/**
	 * Ensures that dependencies have been installed for all package.jsons that have a dependency to
	 * the package described by the passed PackageDescriptor.
	 *
	 * TODO Install just the passed dependency for the package.json
	 */
	private async _ensureDependency(dependency: PackageDescriptor, dependeeName?: string, span = new Span()): Promise<void> {
		await this.dependenciesManager.ensureScanned(span);
		await Promise.all(
			iterate(this.dependenciesManager.packageJsonUris()).map(async uri => {
				try {
					await this.updater.ensure(uri, span);
					const packageJson = JSON.parse(this.inMemoryFileSystem.getContent(uri));
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
		const packageName = getPackageName(originalUri);
		if (!packageName) {
			return originalUri;
		}

		const encodedPackageName = packageName.split('/').map(encodeURIComponent).join('/');

		const packageNameIndex = originalParts.pathname.lastIndexOf('/node_modules/' + encodedPackageName);
		const packageRootUri = url.format({ ...originalParts, pathname: originalParts.pathname.slice(0, packageNameIndex) + `/node_modules/${encodedPackageName}` });
		const packageJsonUri = url.format({ ...originalParts, pathname: originalParts.pathname.slice(0, packageNameIndex) + `/node_modules/${encodedPackageName}/package.json` });

		// Get package.json of dependency
		try {
			await this.updater.ensure(packageJsonUri);
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
			// Special case: @types/ packages are in a subfolder of DefinitelyTyped, named after the package name
			sourcegraphUrl.hash = packageName.substr('@types/'.length) + '/' + sourcegraphUrl.hash;
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

	async textDocumentDefinition(params: TextDocumentPositionParams, span = new Span()): Promise<Location[]> {
		let locations: Location[] = [];
		// First, attempt to get definition before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.dependenciesManager.ensureForFile(params.textDocument.uri, span).catch(err => undefined); // don't wait, but kickoff background job
			locations = await super.textDocumentDefinition(params, span);
		} catch (e) {
			// Ignore
		}
		if (locations.length === 0) {
			await this.dependenciesManager.ensureForFile(params.textDocument.uri, span);
			locations = await super.textDocumentDefinition(params, span);
		}
		await this._rewriteUris(locations);
		return locations;
	}

	/**
	 * This method is the same as textDocument/definition, except that:
	 *
	 * - The method returns metadata about the definition (the same metadata that
	 * workspace/xreferences searches for).
	 * - The concrete location to the definition (location field)
	 * is optional. This is useful because the language server might not be able to resolve a goto
	 * definition request to a concrete location (e.g. due to lack of dependencies) but still may
	 * know some information about it.
	 */
	textDocumentXdefinition(params: TextDocumentPositionParams, span = new Span()): Observable<SymbolLocationInformation[]> {
		// First, attempt to get definition before dependencies fetching is finished.
		this.dependenciesManager.ensureForFile(params.textDocument.uri, span).catch(err => undefined);
		return (super.textDocumentXdefinition(params, span)
			.catch(err => [[]])
			// If no result, wait for dependency installation and retry
			.mergeMap((symbolLocations): Observable<SymbolLocationInformation[]> =>
				symbolLocations.length > 0
					? Observable.of(symbolLocations)
					: Observable.from(this.dependenciesManager.ensureForFile(params.textDocument.uri, span))
						.mergeMap(() => super.textDocumentXdefinition(params, span))
			)
			.mergeAll<any>() as Observable<SymbolLocationInformation>)
			// Add PackageDescriptors to SymbolDescriptor
			.mergeMap((symbolLocation: SymbolLocationInformation) => {
				// If no location is defined, return SymbolLocationInformation unchanged
				if (!symbolLocation.location) {
					return [symbolLocation];
				}
				// Get package name of the dependency in which the symbol is defined in, if any
				const packageName = getPackageName(symbolLocation.location.uri);
				if (packageName) {
					// The symbol is part of a dependency in node_modules
					// Build URI to package.json of the Dependency
					const encodedPackageName = packageName.split('/').map(encodeURIComponent).join('/');
					const parts = url.parse(symbolLocation.location.uri);
					const packageJsonUri = url.format({ ...parts, pathname: parts.pathname!.slice(0, parts.pathname!.lastIndexOf('/node_modules/' + encodedPackageName)) + `/node_modules/${encodedPackageName}/package.json` });
					// Make sure we have the package.json of the dependency available by ensuring the dependency is installed
					return Observable.from(this.dependenciesManager.ensureForFile(packageJsonUri, span))
						// Fetch the package.json of the dependency
						.mergeMap(() => this.updater.ensure(packageJsonUri))
						.map(() => {
							const packageJson: PackageJson = JSON.parse(this.inMemoryFileSystem.getContent(packageJsonUri));
							const { name, version } = packageJson;
							if (name) {
								// Used by the LSP proxy to shortcut database lookup of repo URL for PackageDescriptor
								let repoURL: string | undefined;
								if (name.startsWith('@types/')) {
									// if the dependency package is an @types/ package, point the repo to DefinitelyTyped
									repoURL = 'https://github.com/DefinitelyTyped/DefinitelyTyped';
								} else {
									// else use repository field from package.json
									repoURL = typeof packageJson.repository === 'object' ? packageJson.repository.url : undefined;
								}
								symbolLocation.symbol.package = { name, version, repoURL };
							}
							// Remove location because it points to node_modules instead of the external repo
							symbolLocation.location = undefined;
							return symbolLocation;
						});
				} else {
					// The symbol is defined in the root package of the workspace, not in a dependency
					// Get root package.json
					return Observable.from(this.dependenciesManager.ensureScanned(span))
						.mergeMap(() => this.dependenciesManager.getClosestPackageJson(symbolLocation.location!.uri))
						.map(packageJson => {
							if (!packageJson) {
								// Workspace has no package.json
								return symbolLocation;
							}
							let { name, version } = packageJson;
							if (name) {
								let repoURL = typeof packageJson.repository === 'object' ? packageJson.repository.url : undefined;
								// If the root package is DefinitelyTyped, find out the proper @types package name for each typing
								if (name === 'definitely-typed') {
									// Example:
									// rootUri      file:///
									// symbol URI   file:///node/v6/index.d.ts
									// relative URI        /node/v6/index.d.ts
									// package name         node
									name = '@types/' + decodeURIComponent(urlRelative(this.rootUri, symbolLocation.location!.uri).split('/')[1]);
									version = undefined;
									repoURL = 'https://github.com/DefinitelyTyped/DefinitelyTyped';
								}
								symbolLocation.symbol.package = { name, version, repoURL };
							}
							return symbolLocation;
						});
				}
			})
			// Remove duplicates
			// These can happen if a repository defines the same symbol in multiple locations with
			// interface merging, because we remove the location field
			// See https://github.com/sourcegraph/sourcegraph/issues/5365#issuecomment-294431395
			.distinct(symbolLocation => hashObject(symbolLocation, { respectType: false } as any))
			.toArray();
	}

	async textDocumentHover(params: TextDocumentPositionParams, span = new Span()): Promise<Hover> {
		let hover: Hover = { contents: [] };
		// First, attempt to get hover info before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.dependenciesManager.ensureForFile(params.textDocument.uri, span); // don't wait, but kickoff background job
			hover = await super.textDocumentHover(params, span);
		} catch (e) {
			// Ignore
		}
		if (isEmpty(hover.contents)) {
			await this.dependenciesManager.ensureForFile(params.textDocument.uri, span);
			hover = await super.textDocumentHover(params, span);
		}
		await this._rewriteUris(hover);
		return hover;
	}

	async workspaceSymbol(params: WorkspaceSymbolParams, span = new Span()): Promise<SymbolInformation[]> {
		if (this.dependenciesManager.puntWorkspaceSymbol && (!params.symbol || !params.symbol.package)) {
			throw new Error('workspace/symbol unsupported on this repository');
		}
		return super.workspaceSymbol(params, span);
	}

	/**
	 * The workspace references request is sent from the client to the server to locate project-wide
	 * references to a symbol given its description / metadata.
	 */
	workspaceXreferences(params: WorkspaceReferenceParams, span = new Span()): Observable<ReferenceInformation[]> {
		const dependeePackageName = params.hints ? params.hints.dependeePackageName : undefined;
		const packageDescriptor = params.query.package;
		return ((packageDescriptor ? Observable.from(this._ensureDependency(packageDescriptor, dependeePackageName, span)) : Observable.of(null))
			.mergeMap(() => {
				// Strip the `package` field, because this was not added by the language server
				// TODO is this needed?
				params.query.package = undefined;
				return super.workspaceXreferences(params, span);
			})
			.mergeAll<any>() as Observable<ReferenceInformation>)
			// Add back PackageDescriptors
			.do((referenceInformation: ReferenceInformation) => {
				if (packageDescriptor) {
					referenceInformation.symbol.package = packageDescriptor;
				}
			})
			.toArray();
	}
}
