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

import { TypeScriptService } from 'javascript-typescript-langserver/src/typescript-service';
import { LanguageHandler } from 'javascript-typescript-langserver/src/lang-handler';
import { install, info, infoAlt, parseGitHubInfo } from './yarnshim';
import { FileSystem } from 'javascript-typescript-langserver/src/fs';
import { LayeredFileSystem, LocalRootedFileSystem, walkDirs } from './vfs';
import { uri2path } from 'javascript-typescript-langserver/src/util';
import * as rt from 'javascript-typescript-langserver/src/request-type';

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

	// managedModuleDirs is the set of directories of modules managed
	// by the build handler. It excludes modules already vendored in
	// the repository.
	private managedModuleDirs: Set<string>;
	private managedModuleInit: Map<string, Promise<void>>;

	constructor() {
		this.ls = new TypeScriptService();
		this.managedModuleDirs = new Set<string>();
		this.managedModuleInit = new Map<string, Promise<void>>();
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
				this.managedModuleDirs.add(p);
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

	private async ensureDependenciesForFile(uri: string): Promise<void> {
		if (!this.managedModuleInit) {
			throw new Error("build handler is not yet initialized");
		}
		const p = uri2path(uri);
		for (let d = p; true; d = path.dirname(d)) {
			if (this.managedModuleDirs.has(d)) {
				if (!this.managedModuleInit.has(d)) {
					const installAndRefresher = install(this.remoteFs, d, yarnGlobalDir, path.join(this.yarnOverlayRoot, d)).then(() => {
						return this.ls.projectManager.refreshModuleStructureAt(d);
					}, (err) => {
						this.managedModuleInit.delete(d);
					});
					this.managedModuleInit.set(d, installAndRefresher);
				}
				return this.managedModuleInit.get(d);
			}
			if (path.dirname(d) === d) {
				break;
			}
		}
	}

	private async rewriteURI(uri: string): Promise<{ uri: string, rewritten: boolean }> {
		// if uri is in a dependency module
		const p = uri2path(uri);
		const i = p.indexOf('/node_modules/');
		if (i === -1) {
			return { uri: uri, rewritten: false };
		}

		// if the dependency module is managed by this build handler
		let cwd = p.substring(0, i);
		if (cwd === '') {
			cwd = '/';
		}
		if (!this.managedModuleDirs.has(cwd)) {
			return { uri: uri, rewritten: false };
		}

		// if we can get the module package name heuristically
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

		// if we can fetch the package metadata and if the metadata contains a git URL
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

		// if we can parse the git URL
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
		await this.rewriteURIs(locs)
		return locs;
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

	getWorkspaceSymbols(params: rt.WorkspaceSymbolParamsWithLimit): Promise<SymbolInformation[]> {
		return this.ls.getWorkspaceSymbols(params);
	}

	getDocumentSymbol(params: DocumentSymbolParams): Promise<SymbolInformation[]> {
		return this.ls.getDocumentSymbol(params);
	}

	getWorkspaceReference(params: rt.WorkspaceReferenceParams): Promise<rt.ReferenceInformation[]> {
		return this.ls.getWorkspaceReference(params);
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
