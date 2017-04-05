import * as React from "react";

import URI from "vs/base/common/uri";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { IPosition, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import { Router, __getRouterForWorkbenchOnly, getRevFromRouter } from "sourcegraph/app/router";
import { URIUtils } from "sourcegraph/core/uri";

export interface PathSpec {
	repo: string;
	rev: string | null;
	path: string;
}

export class RouterContext extends React.Component<{}, {}> {
	static childContextTypes: { [key: string]: React.Validator<any> } = {
		router: React.PropTypes.object.isRequired,
	};

	getChildContext(): { router: Router } {
		const router = __getRouterForWorkbenchOnly();
		router.setRouteLeaveHook = () => {
			throw new Error("Cannot set route leave hook outside React Router hierarchy.");
		};
		router.isActive = () => {
			throw new Error("Cannot access isActive outside React Router hierarchy.");
		};
		return { router };
	}

	render(): JSX.Element {
		return this.props.children as JSX.Element;
	}
}

export class Dispatcher<T>{

	private listeners: (((payload: T) => void) | null)[] = [];

	dispatch(payload: T): void {
		Object.freeze(payload);
		this.listeners.forEach((listener) => {
			if (listener) {
				listener(payload);
			}
		});
	}

	subscribe(callback: (payload: Readonly<T>) => void): Disposable {
		const index = this.listeners.length;
		this.listeners.push(callback);
		return {
			dispose: () => {
				this.listeners[index] = null;
			},
		};
	}

}

export class MiniStore<T> extends Dispatcher<T> {

	private initialized: boolean;
	private state: Readonly<T>;

	constructor(initState?: T) {
		super();
		this.subscribe(payload => {
			this.state = payload;
		});
		if (initState) {
			this.dispatch(initState);
			this.initialized = true;
		}
	}

	init(initState: T): void {
		if (this.initialized) {
			throw new Error("store has already been initialized");
		}
		this.dispatch(initState);
		this.initialized = true;
	}

	isInitialized(): boolean {
		return this.initialized;
	}

	getState(): Readonly<T> {
		return this.state;
	}

}

export interface Disposable {
	dispose: () => void;
};

export class Disposables {

	private toDispose: Disposable[] = [];

	public add(d: Disposable): void {
		this.toDispose.push(d);
	}

	public addFn(d: () => void): void {
		this.toDispose.push({
			dispose: d,
		});
	}

	public dispose(): void {
		this.toDispose.forEach(d => {
			d.dispose();
		});
	}

}

export function scrollToLine(editor: ICodeEditor, line: number): void {
	const lineHeight = editor.getConfiguration().lineHeight;
	const scrollPos = line * lineHeight;
	(editor as any)._view.layoutProvider.setScrollPosition({ scrollTop: scrollPos });
}

let contextService: IWorkspaceContextService;
/**
 * setContextService allows a bootstrapped vscode service injector to register
 * its WorkspaceContextService, for use by other functions in this package.
 * It is necessary hack to avoid import cycles.
 */
export function setContextService(service: IWorkspaceContextService): void {
	contextService = service;
}

/**
 * prettifyRev takes a treeish and returns the cosmetic revision, if it is the
 * same as the current workspace revision. This lets us avoid jump to def
 * converting from a nice revision to an absolute commit hash.
 */
export function prettifyRev(newRevision: string | null): string | null {
	const workspace = getCurrentWorkspace();
	const rev = workspace.revState!.commitID;

	if (rev === newRevision) {
		const router = __getRouterForWorkbenchOnly();
		return getRevFromRouter(router) || null;
	}
	return newRevision;
}

export function normalisePosition(model: IReadOnlyModel, position: IPosition): IPosition {
	const word = model.getWordAtPosition(position);
	if (!word) {
		return position;
	}
	// We always hover/j2d on the middle of a word. This is so multiple requests for the same word
	// result in a lookup on the same position.
	return {
		lineNumber: position.lineNumber,
		column: Math.floor((word.startColumn + word.endColumn) / 2),
	};
}

/**
 * getCurrentWorkspace returns the current workspace from the
 * WorkspaceContextService.
 */
export function getCurrentWorkspace(): IWorkspace {
	return contextService.getWorkspace();
}

/**
 * getCurrentWorkspaceRepo returns the current workspace repo from the
 * WorkspaceContextService, e.g. "github.com/gorilla/mux".
 */
export function getCurrentWorkspaceRepo(): string {
	const workspace = getCurrentWorkspace();
	return workspace.resource.authority + workspace.resource.path;
}

/**
 * getCurrentWorkspaceRev returns the current workspace rev from the
 * WorkspaceContextService, if defined.
 */
export function getCurrentWorkspaceRev(): string | null {
	const workspace = getCurrentWorkspace();
	return workspace.revState ? workspace.revState.commitID || null : null;
}

/**
 * getWorkspaceRepoForResource looks for a registered workspace matcing the
 * provided URI in the WorkspaceContextService. It throws an exception if none
 * is found.
 */
export function getWorkspaceRepoForResource(uri: URI): string {
	const registryWorkspace = contextService.tryGetWorkspaceFromRegistry(URIUtils.tryConvertGitToFileURI(uri));
	if (!registryWorkspace) {
		throw new Error(`cannot determine repo URI for resource ${uri}`);
	}
	return registryWorkspace.resource.authority + registryWorkspace.resource.path;
}

/**
 * getWorkspaceRevForResource looks for a registered workspace matcing the
 * provided URI in the WorkspaceContextService. It throws an exception if none
 * is found.
 */
export function getWorkspaceRevForResource(uri: URI): string | null {
	const registryWorkspace = contextService.tryGetWorkspaceFromRegistry(URIUtils.tryConvertGitToFileURI(uri));
	if (!registryWorkspace) {
		throw new Error(`cannot determine repo URI for resource ${uri}`);
	}
	if (registryWorkspace.revState) {
		return registryWorkspace.revState.commitID || null;
	}
	return null;
}

/**
 * getGitBaseResource translates a file-scheme URI into its "base"
 * git-scheme equivalent. The base is determined by the WorkspaceContextService.
 * For a buffer that starts at git commit A then gets a bunch of zap OT ops applied,
 * the transformation will be "file://repo/file.go" -> "git://repo?A#file.go".
 */
export function getGitBaseResource(uri: URI): URI {
	if (uri.scheme !== "file") {
		throw new Error(`invalid uri scheme, expected 'file': ${uri}`);
	}

	const { repo, rev, path } = getURIContext(uri);
	return URI.parse(`git://${repo}?${rev}${path !== "" ? `#${path}` : ""}`);
}

/**
 * getWorkspaceForResource converts a git- or file- scheme URI
 * (for a workspace or document) into the file-scheme workspace
 * for that resource. For a file-scheme workspace, it should be the
 * identity function.
 */
export function getWorkspaceForResource(uri: URI): URI {
	const repo = getWorkspaceRepoForResource(uri);
	return URI.parse(`file://${repo}`);
}

/**
 * isCurrentWorkspace determines if the provided URI matches
 * the URI of the current workspace in WorkspaceContextService.
 */
export function isCurrentWorkspace(uri: URI): boolean {
	return getCurrentWorkspace().resource.toString() === uri.toString();
}

/**
 * isInCurrentWorkspace determines if the provided URI matches
 * the URI or is a subpath of the current workspace in WorkspaceContextService.
 */
export function isInCurrentWorkspace(uri: URI): boolean {
	const workspace = getCurrentWorkspace().resource.toString();
	let uriStr = uri.toString();
	if (uriStr.indexOf(workspace) !== 0) {
		return false;
	}
	uriStr = uriStr.substr(workspace.length);
	if (uriStr.length === 0) {
		// is the workspace
		return true;
	}
	return uriStr[0] === "/";
}

/**
 * getURIContext is used to read the context of a workspace or document URI
 * For a git-scheme URI, these parsed like git://[repo]?[rev]#[path].
 * For a file-scheme URI, they are read from vscode's WorkspaceContextService.
 */
export function getURIContext(uri: URI): { repo: string, path: string, rev: string | null } {
	switch (uri.scheme) {
		case "zap":
		case "git":
			return {
				repo: uri.authority + uri.path,
				rev: uri.query,
				path: decodeURIComponent(uri.fragment)
			};
		case "file":
			const repo = getWorkspaceRepoForResource(uri);
			return {
				repo,
				path: decodeURIComponent(uri.toString().substr(`file://${repo}/`.length)),
				rev: getWorkspaceRevForResource(uri),
			};
	}
	throw new Error(`invalid uri scheme, expected 'git', 'zap' or 'file': ${uri}`);
}
