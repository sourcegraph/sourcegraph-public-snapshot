import * as React from "react";
import { render } from "react-dom";
import { ReferencesWidget } from "sourcegraph/references/ReferencesWidget";
import * as colors from "sourcegraph/util/colors";
import { parseBlob, toBlob } from "sourcegraph/util/url";

const APP_ID = "ref-widget";

export function injectReferencesWidget(): void {
	const widgetContainer = document.getElementById(APP_ID) as HTMLElement;
	if (widgetContainer) {
		widgetContainer.style.backgroundColor = colors.referencesBackgroundColor;
		widgetContainer.style.color = colors.normalFontColor;
		widgetContainer.style.overflowY = "scroll";
		widgetContainer.style.overflowX = "hidden";
		widgetContainer.style.display = window.location.hash.indexOf("$references") === -1 ? "none" : "block";
		widgetContainer.style.borderTop = `1px solid ${colors.borderColor}`;
		widgetContainer.style.zIndex = "1000";

		window.addEventListener("hashchange", (e) => {
			if (e && e.newURL!.indexOf("$references") !== -1) {
				widgetContainer.style.display = "block";

				// References panel was shown, so scroll the blob table so
				// that the selected line is centered.
				const selectedLine = document.querySelector(".code-cell.sg-highlighted")!;
				const blobTable = document.querySelector("#blob-table")!;
				const horizontalScrollBefore = blobTable.scrollLeft;
				selectedLine.scrollIntoView();

				// Ensure horizontal scroll doesn't change.
				blobTable.scrollLeft = horizontalScrollBefore;

				// Center the line on the screen.
				blobTable.scrollTop -= blobTable.getBoundingClientRect().height / 2;
			} else {
			}
		});

		render(<ReferencesWidget onDismiss={dismissReferencesWidget} />, widgetContainer);
	}
}

export function dismissReferencesWidget(): void {
	const widgetContainer = document.getElementById(APP_ID) as HTMLElement;
	widgetContainer.style.display = "none";
	const currURL = parseBlob();
	window.location.href = toBlob({ ...currURL, modal: undefined, modalMode: undefined });
}
