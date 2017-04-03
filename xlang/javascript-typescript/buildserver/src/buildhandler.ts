
import * as rimraf from 'rimraf';
import * as path from 'path';
import {
	TextDocumentPositionParams,
	Location,
	Hover,
	SymbolInformation,
	InitializeResult
} from 'vscode-languageserver';
import { TypeScriptService, TypeScriptServiceOptions } from 'javascript-typescript-langserver/lib/typescript-service';
import { LanguageClientHandler } from 'javascript-typescript-langserver/lib/lang-handler';
import { FileSystem } from 'javascript-typescript-langserver/lib/fs';
import { LayeredFileSystem, LocalRootedFileSystem } from './vfs';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import { CancellationToken, isCancelledError } from 'javascript-typescript-langserver/lib/cancellation';
import {
	PackageDescriptor,
	SymbolLocationInformation,
	WorkspaceReferenceParams,
	WorkspaceSymbolParams,
	ReferenceInformation,
	InitializeParams
} from 'javascript-typescript-langserver/lib/request-type';
import iterate from 'iterare';
import { DependencyManager, PackageJson, getPackageName } from './dependencies';
import * as url from 'url';
import { isEmpty } from 'lodash';
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

export type BuildHandlerFactory = (client: LanguageClientHandler, options: BuildHandlerOptions) => BuildHandler;

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

	constructor(client: LanguageClientHandler, options: BuildHandlerOptions) {
		super(client, options);
	}

	async initialize(params: InitializeParams, token = CancellationToken.None): Promise<InitializeResult> {
		// Workaround for https://github.com/sourcegraph/sourcegraph/issues/4542
		if (params.rootPath && params.rootPath.startsWith('file://')) {
			params.rootPath = uri2path(params.rootPath);
		}
		const result = await super.initialize(params, token);
		this.dependenciesManager = new DependencyManager(this.options.tempDir, this.updater, this.inMemoryFileSystem, this.projectManager, this.logger);
		// Start installation of dependencies in the background
		(async () => {
			try {
				await this.dependenciesManager.ensureScanned();
			} catch (err) {
				if (!isCancelledError(err)) {
					console.error('Dependency initialization failed: ', err);
				}
			}
		})();
		return result;
	}

	/**
	 * Sets up the overlayed file system that includes yarn dependencies
	 */
	protected initializeFileSystems(accessDisk: boolean): void {
		super.initializeFileSystems(accessDisk);
		this.remoteFileSystem = this.fileSystem;
		const overlayFs = new LocalRootedFileSystem(this.root, path.join(this.options.tempDir, 'workspace'));
		this.fileSystem = new LayeredFileSystem([overlayFs, this.remoteFileSystem]);
	}

	async shutdown(): Promise<void> {
		// Delete workspace-specific temporary folder with dependencies
		this.logger.log(`Cleaning up temporary folder ${this.options.tempDir} on shutdown`);
		await new Promise((resolve, reject) => rimraf(this.options.tempDir, err => err ? reject(err) : resolve()));
		await super.shutdown();
	}

	/**
	 * ensureDependenciesToPackage ensures that dependencies have been
	 * installed for all managed module directories that have a
	 * dependency that matches the properties in `pkg`. It does so by
	 * ensuring all dependencies anywhere have been installed. In the
	 * future, this could be optimized by selectively installing
	 * dependencies only for necessary module directories or optimized
	 * even more to install just that dependency in a given managed
	 * module directory.
	 */
	private async ensureDependency(dependency: PackageDescriptor, dependeeName?: string): Promise<void> {
		await this.dependenciesManager.ensureScanned();
		await Promise.all(iterate(this.dependenciesManager.packages).map(([uri, packageJson]): any => {
			if (!dependeeName || packageJson['name'] === dependeeName) {
				return this.dependenciesManager.ensureForFile(uri);
			}
		}));
	}

	/**
	 * Rewrites a given workspace URI to a Sourcegraph `git://repo?rev#path` URI
	 */
	private async rewriteUri(originalUri: string): Promise<string> {
		const originalParts = url.parse(originalUri);

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
			if (!repositoryParts.hostname.endsWith('github.com') || !repositoryParts.pathname) {
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
	private async rewriteUris(result: any): Promise<void> {
		if (Array.isArray(result)) {
			await Promise.all(result.map(element => this.rewriteUris(element)));
		} else if (typeof result === 'object' && result !== null) {
			if (hasUri(result)) {
				result.uri = await this.rewriteUri(result.uri);
			} else {
				await Promise.all(Object.keys(result).map(key => this.rewriteUris(result[key])));
			}
		}
	}

	async getDefinition(params: TextDocumentPositionParams): Promise<Location[]> {
		let locations: Location[];
		// First, attempt to get definition before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.dependenciesManager.ensureForFile(params.textDocument.uri).catch(err => undefined); // don't wait, but kickoff background job
			locations = await super.getDefinition(params);
		} catch (e) { }
		if (!locations || locations.length === 0) {
			await this.dependenciesManager.ensureForFile(params.textDocument.uri);
			await this.projectManager.createConfigurations();
			locations = await super.getDefinition(params);
		}
		await this.rewriteUris(locations);
		return locations;
	}

	async getXdefinition(params: TextDocumentPositionParams): Promise<SymbolLocationInformation[]> {
		let symbolsLocations: SymbolLocationInformation[] = [];
		// First, attempt to get definition before dependencies fetching is finished.
		// If it fails, wait for dependency fetching to finish and then retry.
		try {
			this.dependenciesManager.ensureForFile(params.textDocument.uri).catch(err => undefined);
			symbolsLocations = await super.getXdefinition(params);
		} catch (e) { }
		if (symbolsLocations.length === 0) {
			await this.dependenciesManager.ensureForFile(params.textDocument.uri);
			symbolsLocations = await super.getXdefinition(params);
		}
		// Add PackageDescriptors to SymbolDescriptor
		await Promise.all(symbolsLocations.map(async symbolLocation => {
			// Get package name of the dependency in which the symbol is defined in, if any
			const packageName = getPackageName(symbolLocation.location.uri);
			if (packageName) {
				// The symbol is part of a dependency in node_modules
				// Build URI to package.json of the Dependency
				const encodedPackageName = packageName.split('/').map(encodeURIComponent).join('/');
				const parts = url.parse(symbolLocation.location.uri);
				const packageJsonUri = url.format({ ...parts, pathname: parts.pathname.slice(0, parts.pathname.lastIndexOf('/node_modules/' + encodedPackageName)) + `/node_modules/${encodedPackageName}/package.json` });
				// Make sure we have the package.json of the dependency available by ensuring the dependency is installed
				await this.dependenciesManager.ensureForFile(packageJsonUri);
				// Fetch the package.json of the dependency
				await this.updater.ensure(packageJsonUri);
				const packageJson: PackageJson = JSON.parse(this.inMemoryFileSystem.getContent(packageJsonUri));
				const { name, version } = packageJson;
				if (name) {
					// Used by the LSP proxy to shortcut database lookup of repo URL for PackageDescriptor
					let repoURL: string;
					if (name.startsWith('@types/')) {
						// if the dependency package is an @types/ package, point the repo to DefinitelyTyped
						repoURL = 'https://github.com/DefinitelyTyped/DefinitelyTyped';
					} else {
						// else use repository field from package.json
						repoURL = typeof packageJson.repository === 'object' ? packageJson.repository.url : undefined;
					}
					symbolLocation.symbol.package = { name, version, repoURL };
				}
			} else {
				// The symbol is defined in the root package of the workspace, not in a dependency
				// Get root package.json
				await this.dependenciesManager.ensureScanned();
				const packageJson = this.dependenciesManager.getClosestPackageJson(symbolLocation.location.uri);
				if (!packageJson) {
					// Workspace has no package.json
					return;
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
						name = '@types/' + decodeURIComponent(urlRelative(this.rootUri, symbolLocation.location.uri).split('/')[1]);
						version = undefined;
						repoURL = 'https://github.com/DefinitelyTyped/DefinitelyTyped';
					}
					symbolLocation.symbol.package = { name, version, repoURL };
				}
			}
			// Remove location because ?
			symbolLocation.location = undefined;
		}));
		return symbolsLocations;
	}

	async getHover(params: TextDocumentPositionParams): Promise<Hover> {
		let hover: Hover;
		// First, attempt to get hover info before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.dependenciesManager.ensureForFile(params.textDocument.uri); // don't wait, but kickoff background job
			hover = await super.getHover(params)
		} catch (e) { }
		if (!hover || isEmpty(hover.contents)) {
			await this.dependenciesManager.ensureForFile(params.textDocument.uri);
			hover = await super.getHover(params);
		}
		await this.rewriteUris(hover)
		return hover;
	}

	async getWorkspaceSymbols(params: WorkspaceSymbolParams): Promise<SymbolInformation[]> {
		if (this.dependenciesManager.puntWorkspaceSymbol && (!params.symbol || !params.symbol['package'])) {
			throw new Error("workspace/symbol unsupported on this repository");
		}
		return super.getWorkspaceSymbols(params);
	}

	async getWorkspaceReference(params: WorkspaceReferenceParams): Promise<ReferenceInformation[]> {
		const dependeePackageName = params.hints ? params.hints.dependeePackageName : undefined;
		await this.ensureDependency(params.query.package, dependeePackageName);

		// strip the `package` field, because this was not added by the language server
		const pkgData = params.query.package;
		params.query.package = undefined;

		const refs = await super.getWorkspaceReference(params);

		if (pkgData) {
			for (const ref of refs) {
				ref.symbol.package = pkgData;
			}
		}
		return refs;
	}
}
