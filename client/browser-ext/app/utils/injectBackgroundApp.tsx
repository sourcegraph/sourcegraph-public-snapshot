import * as React from "react";
import { render, unmountComponentAtNode } from "react-dom";

export function injectBackgroundApp(element: JSX.Element | null): void {
	if (document.getElementById("sourcegraph-app-background")) {
		// make this function idempotent
		return;
	}
	const backgroundContainer = document.createElement("div");
	backgroundContainer.id = "sourcegraph-app-background";
	backgroundContainer.style.display = "none";
	document.body.appendChild(backgroundContainer);
	if (element) {
		render(element, backgroundContainer);
	}
}
