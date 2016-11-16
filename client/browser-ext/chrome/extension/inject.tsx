import * as Actions from "../../app/actions";
import {EventLogger} from "../../app/analytics/EventLogger";
import {Background} from "../../app/components/Background";
import {BlobAnnotator} from "../../app/components/BlobAnnotator";
import {configureStore} from "../../app/store/configureStore";
import * as github from "../../app/utils/github";
import {getGitHubRoute, isGitHubURL, parseURL} from "../../app/utils/index";
import * as React from "react";
import {render} from "react-dom";
import {unmountComponentAtNode} from "react-dom";
import {Provider} from "react-redux";

let store = configureStore();

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

	const {repoURI, path, isDelta} = parseURL(window.location);
	if (!repoURI) {
		throw new Error("cannot determine repo URI");
	}

	const files = github.getFileContainers();

	for (let i = 0; i < files.length; ++i) {
		const file = files[i];
		if (github.isInlineCommentContainer(file)) {
			continue;
		}

		const filePath = isDelta ? github.getDeltaFileName(file) : path;
		if (!filePath) {
			throw new Error("cannot infer file path of blob container");
		}

		const blob = file.querySelector(".blob-wrapper") as HTMLElement;
		if (!blob) { // TODO(john): move this logic into BlobAnnotator
			continue;
		}

		const mountEl = document.createElement("div");
		mountEl.className = "btn btn-sm tooltipped tooltipped-n sourcegraph-app-annotator";
		mountEl.style.marginRight = "5px";

		const fileActions = file.querySelector(".file-actions");
		if (!fileActions) {
			throw new Error("cannot locate BlobAnnotator injection site");
		}

		const buttonGroup = fileActions.querySelector(".BtnGroup");
		if (buttonGroup) { // blob view
			// mountEl.style.cssFloat = "none";
			buttonGroup.parentNode.insertBefore(mountEl, buttonGroup);
		} else { // commit & pull request view
			const note = file.querySelector(".show-file-notes");
			if (!note) {
				throw new Error("cannot locate BlobAnnotator injection site");
			}
			mountEl.style.cssFloat = "none";
			note.parentNode.insertBefore(mountEl, note.nextSibling);
		}

		injectComponent(<BlobAnnotator path={filePath} repoURI={repoURI} blobElement={blob} selfElement={mountEl} />, mountEl);
	}
}

function ejectComponent(mountElement: HTMLElement): void {
	unmountComponentAtNode(mountElement);
	mountElement.remove();
}

function injectComponent(component: React.ReactNode, mountElement: HTMLElement): void {
	chrome.runtime.sendMessage({type: "getSessionToken"}, (token) => {
		store.dispatch(Actions.setAccessToken(token));
		render(<Provider store={store}>{component}</Provider>, mountElement);
	});
}

function ejectModules(): void {
	const annotators = document.getElementsByClassName("sourcegraph-app-annotator") as HTMLCollectionOf<HTMLElement>;
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
	injectModules();
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
		chrome.runtime.sendMessage({ type: "setIdentity", identity: ev.detail });
	}
});
