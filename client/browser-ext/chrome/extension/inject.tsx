import * as Actions from "../../app/actions";
import {EventLogger} from "../../app/analytics/EventLogger";
import {Background} from "../../app/components/Background";
import {BlobAnnotator} from "../../app/components/BlobAnnotator";
import {configureStore} from "../../app/store/configureStore";
import * as github from "../../app/utils/github";
import {getGitHubRoute, isGitHubURL, isSourcegraphURL, parseURL} from "../../app/utils/index";
import * as React from "react";
import {render} from "react-dom";
import {unmountComponentAtNode} from "react-dom";
import {Provider} from "react-redux";

let store = configureStore({});

function injectBackgroundApp(): void {
	let backgroundContainer = document.createElement("div");
	backgroundContainer.id = "sourcegraph-app-background";
	backgroundContainer.style.display = "none";
	document.body.appendChild(backgroundContainer);
	injectComponent(<Background />, backgroundContainer);
}

function injectBlobAnnotator(): void {
	if (!isGitHubURL(window.location)) {
		return;
	}

	const {path, isDelta} = parseURL(window.location);
	const files = github.getFileContainers();

	for (let i = 0; i < files.length; ++i) {
		const file = files[i];
		const filePath = isDelta ? github.getDeltaFileName(file) : path;
		if (!filePath) {
			console.error("cannot infer file path of blob container");
		}

		const info = file.querySelector(".file-info");
		const blob = file.querySelector(".blob-wrapper");
		const actn = file.querySelector(".file-actions");
		const note = file.querySelector(".show-file-notes");
		const btng = actn ? actn.querySelector(".BtnGroup") : null;

		if (!blob) {
			continue;
		}

		if (!info) {
			continue;
		}

		const blobAnnotatorContainer = document.createElement("button");
		blobAnnotatorContainer.className = "btn btn-sm tooltipped tooltipped-n sourcegraph-app-annotator";
		blobAnnotatorContainer.style.display = "inline-block";
		blobAnnotatorContainer.style.verticalAlign = "middle";
		blobAnnotatorContainer.style.marginTop = info.tagName === "A" ? "2px" : "0px";
		blobAnnotatorContainer.style.marginRight = info.tagName === "A" && !file.querySelector(".show-outdated-button") ? "0px" : "5px";

		if (actn) {
			if (btng) {
				// TODO(john): remove type cast
				(blobAnnotatorContainer.style as any).float = "none";
				btng.parentNode.insertBefore(blobAnnotatorContainer, btng);
			} else {
				if (note) {
					// TODO(john): remove type cast
					(blobAnnotatorContainer.style as any).float = "none";
					note.parentNode.insertBefore(blobAnnotatorContainer, note.nextSibling);
				} else {
					// TODO(john): remove type cast
					(blobAnnotatorContainer.style as any).float = "left";
					actn.appendChild(blobAnnotatorContainer);
				}
			}
		} else {
			if (info) {
				// TODO(john): remove type cast
				(blobAnnotatorContainer.style as any).float = "right";
				info.parentNode.insertBefore(blobAnnotatorContainer, info.nextSibling);
			} else {
				// TODO(john): remove type cast
				// TODO(john): is this just broken?
				(blobAnnotatorContainer.style as any).float = "left";
				(files as any).appendChild(blobAnnotatorContainer);
			}
		}

		injectComponent(<BlobAnnotator path={filePath} blobElement={blob} infoElement={info} selfElement={blobAnnotatorContainer} />, blobAnnotatorContainer);
	}
}

function ejectComponent(mountElement): void {
	unmountComponentAtNode(mountElement);
	mountElement.remove();
}

function injectComponent(component, mountElement): void {
	render(<Provider store={store}>{component}</Provider>, mountElement);
}

function ejectModules(): void {
	const annotators = document.getElementsByClassName("sourcegraph-app-annotator");
	const background = document.getElementById("sourcegraph-app-background");
	const bootstrap = document.getElementById("sourcegraph-app-bootstrap");

	for (let idx = annotators.length - 1; idx >= 0; idx--) {
		ejectComponent(annotators.item(idx));
	}

	if (background) {
		ejectComponent(background);
	}

	if (bootstrap) {
		bootstrap.remove(); // Not a react component
	}
}

function injectModules(): void {
	if (!document.getElementById("sourcegraph-app-bootstrap")) {
		injectBackgroundApp();
		injectBlobAnnotator();

		// Add invisible div to the page to indicate injection has completed.
		let el = document.createElement("div");
		el.id = "sourcegraph-app-bootstrap";
		el.style.display = "none";
		document.body.appendChild(el);
	}
}

window.addEventListener("load", () => {
	if (isSourcegraphURL(window.location)) {
		injectModules();
	} else if (isGitHubURL(window.location)) {
		chrome.runtime.sendMessage({type: "getSessionToken"}, {}, (token) => {
			store.dispatch(Actions.setAccessToken(token));
			injectModules();
		});
	}
	chrome.runtime.sendMessage({type: "getIdentity"}, {}, (identity) => {
		if (identity) {
			EventLogger.updatePropsForUser(identity);
		}
	});
});

document.addEventListener("keydown", (e: KeyboardEvent) => {
	if (getGitHubRoute(window.location) !== "blob") {
		return;
	}
	if ((e.target as HTMLElement).tagName === "INPUT" ||
		(e.target as HTMLElement).tagName === "SELECT" ||
		(e.target as HTMLElement).tagName === "TEXTAREA") {
			return;
	}

	if (e.keyCode === 85) {
		const annButtons = document.getElementsByClassName("sourcegraph-app-annotator");
		if (annButtons.length === 1) {
			const annButtonA = annButtons[0].getElementsByTagName("A");
			if (annButtonA.length === 1 && (annButtonA[0] as any).href) {
				window.open((annButtonA[0] as any).href, "_blank");
			}
		}
	}
});

document.addEventListener("pjax:end", () => {
	// Unmount and remount react components because pjax breaks
	// dynamically registered event handlers like mouseover/click etc..
	ejectModules();
	injectModules();
});

document.addEventListener("sourcegraph:identify", (ev: CustomEvent) => {
	if (ev && ev.detail) {
		EventLogger.updatePropsForUser(ev.detail);
		chrome.runtime.sendMessage({type: "setIdentity", identity: ev.detail}, {});
	}
});
