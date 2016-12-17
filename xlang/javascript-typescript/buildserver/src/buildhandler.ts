/// <reference path="../node_modules/vscode/thenable.d.ts" />

import * as fs from 'fs';
import * as rimraf from 'rimraf';

import {
	IConnection,
	createConnection,
	InitializeParams,
	InitializeResult,
	TextDocuments,
	TextDocumentPositionParams,
	TextDocumentSyncKind,
	Definition,
	ReferenceParams,
	Position,
	Location,
	Hover,
	WorkspaceSymbolParams,
	DocumentSymbolParams,
	SymbolInformation,
	RequestType,
	Range,
	DidOpenTextDocumentParams,
	DidCloseTextDocumentParams,
	DidChangeTextDocumentParams,
	DidSaveTextDocumentParams
} from 'vscode-languageserver';

import { TypeScriptService } from 'javascript-typescript-langserver/src/typescript-service';
import { LanguageHandler } from 'javascript-typescript-langserver/src/lang-handler';
import * as rt from 'javascript-typescript-langserver/src/request-type';
import { install, info, infoAlt } from './yarnshim';
import { FileSystem, RemoteFileSystem } from 'javascript-typescript-langserver/src/fs';
import { LayeredFileSystem, LocalRootedFileSystem, walkDirs } from './vfs';
import * as temp from 'temp';
import * as path from 'path';
import { uri2path } from 'javascript-typescript-langserver/src/util';

interface HasURI {
	uri: string;
}

const yarnGlobalDir = "/tmp/tsjs-yarn-global/";

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

	private moduleDirs: Set<string>;
	private moduleInit: Map<string, Promise<void>>;

	constructor() {
		this.ls = new TypeScriptService();
		this.moduleDirs = new Set<string>();
		this.moduleInit = new Map<string, Promise<void>>();
	}

	async initialize(params: InitializeParams, remoteFs: FileSystem, strict: boolean): Promise<InitializeResult> {
		const yarndir = await new Promise<string>((resolve, reject) => {
			temp.mkdir("tsjs-yarn", (err: any, dirPath: string) => {
				if (err) {
					return reject(err);
				}
				return resolve(dirPath);
			});
		});
		this.yarndir = yarndir;
		this.yarnOverlayRoot = path.join(yarndir, "workspace");
		this.remoteFs = remoteFs;

		await walkDirs(remoteFs, "/", async (p, entries) => {
			for (const entry of entries) {
				if (!entry.dir && entry.name === "package.json") {
					this.moduleDirs.add(p);
				}
			}
		});

		const overlayFs = new LocalRootedFileSystem(this.yarnOverlayRoot);
		const lsfs = new LayeredFileSystem([overlayFs, remoteFs]);
		this.lsfs = lsfs;

		return this.ls.initialize(params, lsfs, strict);
	}

	shutdown(): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			rimraf(this.yarndir, (err) => {
				if (err) {
					return reject(err);
				} else {
					return resolve();
				}
			});
		});
	}

	private async ensureDependenciesForFile(uri: string): Promise<void> {
		if (!this.moduleInit) {
			throw new Error("build handler is not yet initialized");
		}
		const p = uri2path(uri);
		for (let d = p; true; d = path.dirname(d)) {
			if (this.moduleDirs.has(d)) {
				if (!this.moduleInit.has(d)) {
					const installAndRefresher = install(this.remoteFs, d, yarnGlobalDir, path.join(this.yarnOverlayRoot, d)).then(() => {
						return this.ls.projectManager.refreshModuleStructureAt(d);
					});
					this.moduleInit.set(d, installAndRefresher);
				}
				return this.moduleInit.get(d);
			}
			if (path.dirname(d) === d) {
				break;
			}
		}
		return Promise.resolve();
	}

	/*
	 * rewriteURIs is a kludge until we have textDocument/xdefinition.
	 */
	private async rewriteURIs(result: any): Promise<void> {
		if (!result) {
			return Promise.resolve();
		}

		if ((<HasURI>result).uri) {
			const p = uri2path(result.uri);
			const i = p.indexOf('/node_modules/');
			if (i !== -1) {
				let cwd = p.substring(0, i);
				if (cwd === '') {
					cwd = '/';
				}
				const cmp = p.substr(i + '/node_modules/'.length).split('/');
				const subpath = path.posix.join(...cmp.slice(1));
				let pkg: string;
				if (cmp.length >= 2 && cmp[0] === "@types") {
					pkg = cmp[0] + "/" + cmp[1];
				} else if (cmp.length >= 1) {
					pkg = cmp[0];
				}

				if (pkg) {
					try {
						const pkginfo = await info(cwd, yarnGlobalDir, path.join(this.yarnOverlayRoot, cwd), pkg);
						if (pkginfo.repository && pkginfo.repository.url && pkginfo.repository.type === 'git') {
							const pkgUri = cleanGitUrl(pkginfo.repository.url);
							const pkgHead = pkginfo.gitHead;
							result.uri = makeUri(pkgUri, pkgHead, subpath);
						}
					} catch (e) {
						try {
							const pkginfo = await infoAlt(this.remoteFs, cwd, yarnGlobalDir, path.join(this.yarnOverlayRoot, cwd), pkg);
							result.uri = makeUri(pkginfo.repository.url, pkginfo.gitHead, subpath);
						} catch (e) {
							console.error("could not rewrite dependency uri,", result.uri, ", due to error:", e);
						}
					}
				}
			}
		}

		if (Array.isArray(result)) {
			for (const e of result) {
				await this.rewriteURIs(e);
			}
			return Promise.resolve();
		}
		return Promise.resolve();
	}

	async getDefinition(params: TextDocumentPositionParams): Promise<Location[]> {
		let locs: Location[];
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
		let hover: Hover;
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

function cleanGitUrl(url: string): string {
	if (url.startsWith("git+https://")) {
		url = url.substr("git+https://".length);
	}
	if (url.startsWith("https://")) {
		url = url.substr("https://".length);
	}
	if (url.startsWith("www.")) {
		url = url.substr("www.".length);
	}
	if (!url.startsWith("git://")) {
		url = "git://" + url;
	}

	if (url.endsWith(".git")) {
		url = url.substr(0, url.length - ".git".length);
	}
	return url;
}

function makeUri(repoUri: string, version: string | null, path: string): string {
	const versionPart = version ? "?" + version : "";
	if (path.startsWith("/")) {
		path = path.substr(1);
	}
	return repoUri + versionPart + "#" + path;
}
