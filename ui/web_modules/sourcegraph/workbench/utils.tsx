import * as React from "react";

import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { IPosition, IReadOnlyModel } from "vs/editor/common/editorCommon";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import { Router, __getRouterForWorkbenchOnly, getRevFromRouter } from "sourcegraph/app/router";
import { URIUtils } from "sourcegraph/core/uri";
import { Services } from "sourcegraph/workbench/services";

export const NoopDisposer = { dispose: () => {/* */ } };

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

export class MiniStore<T>{

	private listeners: (((payload: T) => void) | null)[] = [];

	dispatch(payload: T): void {
		this.listeners.forEach((listener) => {
			if (listener) {
				listener(payload);
			}
		});
	}

	subscribe(callback: (payload: T) => void): Disposable {
		const index = this.listeners.length;
		this.listeners.push(callback);
		return {
			dispose: () => {
				this.listeners[index] = null;
			},
		};
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

/**
 * prettifyRev takes a treeish and returns the cosmetic revision, if it is the
 * same as the current workspace revision. This lets us avoid jump to def
 * converting from a nice revision to an absolute commit hash.
 */
export function prettifyRev(newRevision: string | null): string | null {
	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	const workspace = workspaceService.getWorkspace();
	const { rev } = URIUtils.repoParams(workspace.resource);

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
