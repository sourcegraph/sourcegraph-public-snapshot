import { ReferencesWidget } from "app/references/ReferencesWidget";
import { store } from "app/references/store";
import * as React from "react";
import { render } from "react-dom";
import * as colors from "app/util/colors";

const APP_ID = "sourcegraph-references-widget";

function createAppContainerIfNotExists(tag: string): HTMLElement | undefined {
	if (document.getElementById(APP_ID)) {
		return;
	}
	const el = document.createElement(tag);
	el.id = APP_ID;
	return el;
}

export function injectReferencesWidget(): void {
	const widgetContainer = createAppContainerIfNotExists("div") as HTMLElement;
	if (widgetContainer) {
		widgetContainer.style.position = "fixed";
		widgetContainer.style.backgroundColor = colors.referencesBackgroundColor;
		widgetContainer.style.width = "100%";
		widgetContainer.style.height = "350px";
		widgetContainer.style.left = "0px";
		widgetContainer.style.top = `calc(100vh - 350px)`;
		widgetContainer.style.visibility = "hidden";
		widgetContainer.style.overflow = "auto";
		widgetContainer.style.borderTop = `1px solid ${colors.borderColor}`;
		widgetContainer.style.zIndex = "1000";
		document.body.appendChild(widgetContainer);

		// Handle scrolling ourselves so that scrolling inside the widget does
		// not cause the page to start scrolling (which is a very jarring
		// experience).
		widgetContainer.addEventListener("wheel", (e: WheelEvent) => {
			e.preventDefault();
			widgetContainer.scrollTop += e.deltaY;
		});

		store.subscribe((state) => {
			widgetContainer.style.visibility = state.docked ? "visible" : "hidden";
		});

		render(<ReferencesWidget />, widgetContainer);
	}
}
