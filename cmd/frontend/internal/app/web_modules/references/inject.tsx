import { ReferencesWidget } from "app/references/ReferencesWidget";
import * as colors from "app/util/colors";
import * as React from "react";
import { render } from "react-dom";

const APP_ID = "ref-widget";

export function injectReferencesWidget(): void {
	const widgetContainer = document.getElementById(APP_ID) as HTMLElement;
	if (widgetContainer) {
		widgetContainer.style.backgroundColor = colors.referencesBackgroundColor;
		widgetContainer.style.color = colors.normalFontColor;
		widgetContainer.style.overflow = "auto";
		widgetContainer.style.display = window.location.hash.indexOf("$references") === -1 ? "none" : "block";
		widgetContainer.style.borderTop = `1px solid ${colors.borderColor}`;
		widgetContainer.style.zIndex = "1000";

		window.addEventListener("hashchange", (e) => {
			if (e && e.newURL!.indexOf("$references") !== -1) {
				widgetContainer.style.display = "block";
			}
		});

		render(<ReferencesWidget onDismiss={() => {
			widgetContainer.style.display = "none";
		}} />, widgetContainer);
	}
}
