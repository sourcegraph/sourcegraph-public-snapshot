import { ReferencesWidget } from "app/references/ReferencesWidget";
import { store } from "app/references/store";
import * as React from "react";
import { render } from "react-dom";

const APP_ID = "sourcegraph-references-widget";

function createAppContainerIfNotExists(tag: string): HTMLElement | undefined {
	if (document.getElementById(APP_ID)) {
		return;
	}
	const el = document.createElement(tag);
	el.id = APP_ID;
	return el;
}

export function injectGitHub(): void {
	const widgetContainer = createAppContainerIfNotExists("div") as HTMLElement;
	if (widgetContainer) {
		widgetContainer.style.position = "fixed";
		widgetContainer.style.backgroundColor = "#fafbfc";
		widgetContainer.style.width = "100%";
		widgetContainer.style.height = "350px";
		widgetContainer.style.left = "0px";
		widgetContainer.style.top = `calc(100vh - 350px)`;
		widgetContainer.style.visibility = "hidden";
		widgetContainer.style.overflow = "auto";
		widgetContainer.style.borderTop = "1px solid #e1e4e8";
		document.body.appendChild(widgetContainer);

		store.subscribe((state) => {
			widgetContainer.style.visibility = state.docked ? "visible" : "hidden";
		});

		render(<ReferencesWidget />, widgetContainer);
	}
}
