import * as React from "react";

import URI from "vs/base/common/uri";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { IEditorInput } from "vs/platform/editor/common/editor";

import { Router, __getRouterForWorkbenchOnly } from "sourcegraph/app/router";

export function getResource(input: IEditorInput): URI {
	if (input["resource"]) {
		return (input as any).resource;
	} else {
		throw new Error("Couldn't find resource.");
	}
}

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
