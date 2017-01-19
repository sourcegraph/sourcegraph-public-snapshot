/// <reference path="../node_modules/vscode/thenable.d.ts" />

import * as rimraf from 'rimraf';
import * as temp from 'temp';
import * as path from 'path';
import * as os from 'os';

import {
	InitializeParams,
	InitializeResult,
	TextDocumentPositionParams,
	ReferenceParams,
	Location,
	Hover,
	DocumentSymbolParams,
	SymbolInformation,
	DidOpenTextDocumentParams,
	DidCloseTextDocumentParams,
	DidChangeTextDocumentParams,
	DidSaveTextDocumentParams
} from 'vscode-languageserver';

import { TypeScriptService } from 'javascript-typescript-langserver/lib/typescript-service';
import { LanguageHandler } from 'javascript-typescript-langserver/lib/lang-handler';
import { install, info, infoAlt, parseGitHubInfo } from './yarnshim';
import { FileSystem } from 'javascript-typescript-langserver/lib/fs';
import { LayeredFileSystem, LocalRootedFileSystem, walkDirs, readFile } from './vfs';
import { uri2path } from 'javascript-typescript-langserver/lib/util';
import * as rt from 'javascript-typescript-langserver/lib/request-type';

interface HasURI {
	uri: string;
}

const yarnGlobalDir = path.join(os.tmpdir(), "tsjs-yarn-global");

console.error("Using", yarnGlobalDir, "as yarn global directory");

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
export class BuildHandler implements LanguageHandler {
	private remoteFs: FileSystem;
	private ls: TypeScriptService;
	private lsfs: FileSystem;
	private yarndir: string;
	private yarnOverlayRoot: string;

	/**
	 * managedModuleConfig maps from directory to configuration for
	 * each module managed by the build handler. It excludes modules
	 * already vendored in the repository.
	 */
	private managedModuleConfig = new Map<string, any>();
	private managedModuleInit = new Map<string, Promise<Map<string, rt.DependencyReference>>>();
	private puntWorkspaceSymbol = false;

	constructor() {
		this.ls = new TypeScriptService();
	}

	async initialize(params: InitializeParams, remoteFs: FileSystem, strict: boolean): Promise<InitializeResult> {
		const yarndir = await new Promise<string>((resolve, reject) => {
			temp.mkdir("tsjs-yarn", (err: any, dirPath: string) => err ? reject(err) : resolve(dirPath));
		});
		this.yarndir = yarndir;
		this.yarnOverlayRoot = path.join(yarndir, "workspace");
		this.remoteFs = remoteFs;

		await walkDirs(remoteFs, "/", async (p, entries) => {
			let foundPackageJson = false;
			let foundModulesDir = false;
			for (const entry of entries) {
				if (!entry.dir && entry.name === "package.json") {
					foundPackageJson = true;
				}
				if (entry.dir && entry.name === "node_modules") {
					foundModulesDir = true
				}
			}
			if (foundPackageJson && !foundModulesDir) {
				const config = JSON.parse(await readFile(remoteFs, path.join(p, 'package.json')));
				if (config['name'] === 'definitely-typed') {
					this.puntWorkspaceSymbol = true;
				}
				this.managedModuleConfig.set(p, config);
			}
		});

		const overlayFs = new LocalRootedFileSystem(this.yarnOverlayRoot);
		const lsfs = new LayeredFileSystem([overlayFs, remoteFs]);
		this.lsfs = lsfs;

		return this.ls.initialize(params, lsfs, strict);
	}

	shutdown(): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			rimraf(this.yarndir, (err) => err ? reject(err) : resolve());
		});
	}

	private getManagedModuleDir(uri: string): string | null {
		const p = uri2path(uri);
		for (let d = p; true; d = path.dirname(d)) {
			if (this.managedModuleConfig.has(d)) {
				return d;
			}

			if (path.dirname(d) === d) {
				break;
			}
		}
		return null;
	}

	private async ensureDependenciesForFile(uri: string): Promise<void> {
		if (!this.managedModuleInit) {
			throw new Error("build handler is not yet initialized");
		}
		const d = this.getManagedModuleDir(uri);
		if (!d) {
			return;
		}

		await this.initManagedModule(d);
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
	private async ensureDependency(dependency: rt.PackageDescriptor, dependeeName?: string): Promise<void> {
		if (!this.managedModuleInit) {
			throw new Error("build handler is not yet initialized");
		}
		await Promise.all(Array.from(this.managedModuleConfig.keys(), (d) => {
			const config = this.managedModuleConfig.get(d);
			if (!dependeeName || config['name'] === dependeeName) {
				return this.initManagedModule(d);
			} else {
				return Promise.resolve();
			}
		}));
	}

	private async initManagedModule(dir: string): Promise<void> {
		let ready = this.managedModuleInit.get(dir);
		if (!ready) {
			ready = install(this.remoteFs, dir, yarnGlobalDir, path.join(this.yarnOverlayRoot, dir)).then(async (pathToDep) => {
				await this.ls.projectManager.refreshModuleStructureAt(dir);
				return pathToDep;
			}, (err) => {
				this.managedModuleInit.delete(dir);
			});
			this.managedModuleInit.set(dir, ready);
		}
		await ready;
	}

	private async rewriteURI(uri: string): Promise<{ uri: string, rewritten: boolean }> {
		// if uri is not in a dependency module, return untouched
		const p = uri2path(uri);
		const i = p.indexOf('/node_modules/');
		if (i === -1) {
			return { uri: uri, rewritten: false };
		}

		// if the dependency module is not managed by this build handler
		let cwd = p.substring(0, i);
		if (cwd === '') {
			cwd = '/';
		}
		if (!this.managedModuleConfig.has(cwd)) {
			return { uri: uri, rewritten: false };
		}

		// get the module package name heuristically, otherwise punt
		const cmp = p.substr(i + '/node_modules/'.length).split('/');
		const subpath = path.posix.join(...cmp.slice(1));
		let pkg: string | undefined;
		if (cmp.length >= 2 && cmp[0] === "@types") {
			pkg = cmp[0] + "/" + cmp[1];
		} else if (cmp.length >= 1) {
			pkg = cmp[0];
		}
		if (!pkg) {
			return { uri: uri, rewritten: false };
		}

		// fetch the package metadata and extract the git URL from metadata if it exists; otherwise punt
		let pkginfo;
		try {
			pkginfo = await info(cwd, yarnGlobalDir, path.join(this.yarnOverlayRoot, cwd), pkg);
		} catch (e) { }
		if (!pkginfo) {
			try {
				pkginfo = await infoAlt(this.remoteFs, cwd, yarnGlobalDir, path.join(this.yarnOverlayRoot, cwd), pkg);
			} catch (e) {
				console.error("could not rewrite dependency uri,", uri, ", due to error:", e);
				return { uri: uri, rewritten: false };
			}
		}
		if (!pkginfo.repository || !pkginfo.repository.url || pkginfo.repository.type !== 'git') {
			return { uri: uri, rewritten: false };
		}

		// parse the git URL if possible, otherwise punt
		const pkgUrlInfo = parseGitHubInfo(pkginfo.repository.url);
		if (!pkgUrlInfo || !pkgUrlInfo.repository) {
			return { uri: uri, rewritten: false };
		}

		return { uri: makeUri(pkgUrlInfo.repository.url, subpath, pkginfo.gitHead), rewritten: true };
	}

	/*
	 * rewriteURIs is a kludge until we have textDocument/xdefinition.
	 */
	private async rewriteURIs(result: any): Promise<void> {
		if (!result) {
			return;
		}

		if ((<HasURI>result).uri) {
			const { uri, rewritten } = await this.rewriteURI(result.uri);
			if (rewritten) {
				result.uri = uri;
			}
		}

		if (Array.isArray(result)) {
			for (const e of result) {
				await this.rewriteURIs(e);
			}
		} else if (typeof result === "object") {
			for (const k in result) {
				await this.rewriteURIs(result[k]);
			}
		}
	}

	async getDefinition(params: TextDocumentPositionParams): Promise<Location[]> {
		let locs: Location[] = [];
		// First, attempt to get definition before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.ensureDependenciesForFile(params.textDocument.uri); // don't wait, but kickoff background job
			locs = await this.ls.getDefinition(params);
		} catch (e) { }
		if (!locs || locs.length === 0) {
			await this.ensureDependenciesForFile(params.textDocument.uri);
			locs = await this.ls.getDefinition(params);
		}
		await this.rewriteURIs(locs);
		return locs;
	}

	async getXdefinition(params: TextDocumentPositionParams): Promise<rt.SymbolLocationInformation[]> {
		let syms: rt.SymbolLocationInformation[] = [];
		// First, attempt to get definition before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.ensureDependenciesForFile(params.textDocument.uri); // don't wait, but kickoff background job
			syms = await this.ls.getXdefinition(params);
		} catch (e) { }
		if (!syms || syms.length === 0) {
			await this.ensureDependenciesForFile(params.textDocument.uri);
			syms = await this.ls.getXdefinition(params);
		}

		// For symbols defined in dependencies, remove the location field and add in dependency package metadata.
		await Promise.all(syms.map((sym) => this.rewriteSymbol(sym)));

		await this.rewriteURIs(syms);
		return syms;
	}

	private async rewriteSymbol(sym: rt.SymbolLocationInformation): Promise<void> {
		const dep = await this.getDepContainingSymbol(sym);
		if (!dep) {
			let moduleDir = this.getManagedModuleDir(sym.location.uri);
			if (!moduleDir) {
				return;
			}
			const pkgJson = JSON.parse(this.ls.projectManager.getFs().readFile(path.join(moduleDir, "package.json")));
			let name = pkgJson['name'];
			let version = pkgJson['version'];
			let repoURL = pkgJson['repository'] ? pkgJson['repository']['url'] : undefined;
			if (name === 'definitely-typed') { // special case DefinitelyTyped
				name = "@types/" + uri2path(sym.location.uri).split(path.sep)[1];
				version = undefined;
				repoURL = 'https://github.com/DefinitelyTyped/DefinitelyTyped';
			}
			if (name) {
				sym.symbol.package = { name, version, repoURL };
			}
			return;
		}

		sym.symbol.package = dep.attributes;
		const pkgName = sym.symbol.package.name;
		if (pkgName && pkgName.startsWith('@types/')) {
			sym.symbol.package.repoURL = 'https://github.com/DefinitelyTyped/DefinitelyTyped';
		}
		sym.location = undefined;
	}

	/**
	 * getDepContainingSymbol returns the dependency that contains the
	 * symbol or null if the symbol is not defined in a dependency.
	 */
	private async getDepContainingSymbol(sym: rt.SymbolLocationInformation): Promise<rt.DependencyReference | null> {
		const moduledir = this.getManagedModuleDir(sym.location.uri);
		if (!moduledir) {
			return null;
		}
		const pathToDep = await this.managedModuleInit.get(moduledir);
		const p = uri2path(sym.location.uri);
		for (let d = p; true; d = path.dirname(d)) {
			if (pathToDep.has(d)) {
				return pathToDep.get(d);
			}

			if (path.dirname(d) === d) {
				break;
			}
		}
		return null;
	}

	async getHover(params: TextDocumentPositionParams): Promise<Hover> {
		let hover: Hover | null = null;
		// First, attempt to get hover info before dependencies
		// fetching is finished. If it fails, wait for dependency
		// fetching to finish and then retry.
		try {
			this.ensureDependenciesForFile(params.textDocument.uri); // don't wait, but kickoff background job
			hover = await this.ls.getHover(params)
		} catch (e) { }
		if (!hover) {
			await this.ensureDependenciesForFile(params.textDocument.uri);
			hover = await this.ls.getHover(params);
		}
		await this.rewriteURIs(hover)
		return hover;
	}

	getReferences(params: ReferenceParams): Promise<Location[]> {
		return this.ls.getReferences(params);
	}

	getDependencies(): Promise<rt.DependencyReference[]> {
		return this.ls.getDependencies();
	}

	getPackages(): Promise<rt.PackageInformation[]> {
		return this.ls.getPackages();
	}

	async getWorkspaceSymbols(params: rt.WorkspaceSymbolParams): Promise<SymbolInformation[]> {
		if (this.puntWorkspaceSymbol) {
			return Promise.reject("workspace/symbol unsupported on this repository");
		}
		return this.ls.getWorkspaceSymbols(params);
	}

	getDocumentSymbol(params: DocumentSymbolParams): Promise<SymbolInformation[]> {
		return this.ls.getDocumentSymbol(params);
	}

	async getWorkspaceReference(params: rt.WorkspaceReferenceParams): Promise<rt.ReferenceInformation[]> {
		const dependeePackageName = params.hints ? params.hints.dependeePackageName : undefined;
		await this.ensureDependency(params.query.package, dependeePackageName);

		// strip the `package` field, because this was not added by the language server
		const pkgData = params.query.package;
		params.query.package = undefined;

		const refs = await this.ls.getWorkspaceReference(params);

		for (const ref of refs) {
			ref.symbol.package = pkgData;
		}
		return refs;
	}

	didOpen(params: DidOpenTextDocumentParams) {
		return this.ls.didOpen(params);
	}

	didChange(params: DidChangeTextDocumentParams) {
		return this.ls.didChange(params);
	}

	didClose(params: DidCloseTextDocumentParams) {
		return this.ls.didClose(params);
	}

	didSave(params: DidSaveTextDocumentParams) {
		return this.ls.didSave(params);
	}
}

function makeUri(repoUri: string, path: string, version?: string): string {
	const versionPart = version ? "?" + version : "";
	if (path.startsWith("/")) {
		path = path.substr(1);
	}
	return repoUri + versionPart + "#" + path;
}
