import * as React from "react";
import { render } from "react-dom";
import { EditorApp } from "./EditorApp";

const APP_ID = "sourcegraph-connect-editor";

function createAppContainerIfNotExists(tag: string): HTMLElement | undefined {
	if (document.getElementById(APP_ID)) {
		return;
	}
	const el = document.createElement(tag);
	el.id = APP_ID;
	return el;
}

export function injectSourcegraph(): void {
	if (process.env.NODE_ENV !== "development") {
		return;
	}
	setTimeout(() => {
		const app = createAppContainerIfNotExists("div") as HTMLElement;
		app.className = "actionbar-link";
		if (app) {
			let actionbar: any = document.querySelector(".part.actionbar");
			if (actionbar) {
				actionbar = actionbar.querySelector(".actionbar");
			}
			if (actionbar) {
				actionbar.insertBefore(app, actionbar.firstChild);
			}
			render(<EditorApp />, app);
		}
	}, 1000);
}

export function injectGitHub(): void {
	if (process.env.NODE_ENV !== "development") {
		return;
	}
	const app = createAppContainerIfNotExists("li");
	if (app) {
		const pageheadActions = document.querySelector(".pagehead-actions");
		if (pageheadActions) {
			pageheadActions.insertBefore(app, pageheadActions.children[0]);
			render(<EditorApp />, app);
		}
	}
}
