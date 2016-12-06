import { useAccessToken } from "../../app/backend/xhr";
import { Background } from "../../app/components/Background";
import { BlobAnnotator } from "../../app/components/BlobAnnotator";
import { EventLogger } from "../../app/utils/EventLogger";
import * as github from "../../app/utils/github";
import { getGitHubRoute, isGitHubURL, parseURL } from "../../app/utils/index";
import * as React from "react";
import { render, unmountComponentAtNode } from "react-dom";

function ejectComponent(mount: HTMLElement): void {
	try {
		unmountComponentAtNode(mount);
		mount.remove();
	} catch (e) {
		console.error(e);
	}
}

// NOTE: injectModules is idempotent, so safe to call multiple times on the same page.
function injectModules(): void {
	chrome.runtime.sendMessage({ type: "getSessionToken" }, (token) => {
		if (token) {
			useAccessToken(token);
		}
		injectBackgroundApp();
		injectBlobAnnotators();
	});
}

function injectBackgroundApp(): void {
	if (document.getElementById("sourcegraph-app-background")) {
		// make this function idempotent
		return;
	}

	let backgroundContainer = document.createElement("div");
	backgroundContainer.id = "sourcegraph-app-background";
	backgroundContainer.style.display = "none";
	document.body.appendChild(backgroundContainer);
	render(<Background />, backgroundContainer);
}

function injectBlobAnnotators(): void {
	if (!isGitHubURL(window.location)) {
		return;
	}

	const {repoURI, path, isDelta} = parseURL(window.location);
	if (!repoURI) {
		console.error("cannot determine repo URI");
		return;
	}

	const uri = repoURI;
	function addBlobAnnotator(file: HTMLElement, mount: HTMLElement): void {
		const filePath = isDelta ? github.getDeltaFileName(file) : path;
		if (!filePath) {
			console.error("cannot determine file path");
			return;
		}

		const blob = github.tryGetBlobElement(file);
		if (!blob) {
			// File contents are collapsed (e.g. for large files).
			// Asynchronously apply annotator after the user clicks on the file (wait for file contents to load).
			file.addEventListener("click", () => setTimeout(() => addBlobAnnotator(file, mount), 2500));
			return;
		}

		if (file.className.includes("sg-blob-annotated")) {
			// make this function idempotent
			return;
		}
		file.className = `${file.className} sg-blob-annotated`;
		render(<BlobAnnotator path={filePath} repoURI={uri} blobElement={blob} />, mount);
	}

	const files = github.getFileContainers();
	for (let i = 0; i < files.length; ++i) {
		const file = files[i];
		const mount = github.createBlobAnnotatorMount(file);
		if (!mount) {
			return;
		}
		addBlobAnnotator(file, mount);
	}
}

function ejectModules(): void {
	const annotators = document.getElementsByClassName("sourcegraph-app-annotator") as HTMLCollectionOf<HTMLElement>;
	const background = document.getElementById("sourcegraph-app-background");

	for (let idx = annotators.length - 1; idx >= 0; idx--) {
		ejectComponent(annotators.item(idx));
	}

	if (background) {
		ejectComponent(background);
	}

	const annotated = document.getElementsByClassName("sg-blob-annotated") as HTMLCollectionOf<HTMLElement>;
	for (let idx = annotated.length - 1; idx >= 0; idx--) {
		annotated.item(idx).remove();
	}
}

window.addEventListener("load", () => {
	injectModules();
	chrome.runtime.sendMessage({ type: "getIdentity" }, (identity) => {
		if (identity) {
			EventLogger.updatePropsForUser(identity);
		}
	});
	setTimeout(injectModules, 5000); // extra data may be loaded asynchronously; reapply after timeout
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
	setTimeout(injectModules, 5000); // extra data may be loaded asynchronously; reapply after timeout
});

document.addEventListener("sourcegraph:identify", (ev: CustomEvent) => {
	if (ev && ev.detail) {
		EventLogger.updatePropsForUser(ev.detail);
		chrome.runtime.sendMessage({ type: "setIdentity", identity: ev.detail });
	} else {
		console.error("sourcegraph:identify missing details");
	}
});
