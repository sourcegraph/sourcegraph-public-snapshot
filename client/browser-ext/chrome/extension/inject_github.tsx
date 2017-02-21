import * as React from "react";
import { render, unmountComponentAtNode } from "react-dom";
import { useAccessToken } from "../../app/backend/xhr";
import { BlobAnnotator } from "../../app/components/BlobAnnotator";
import { GitHubBackground } from "../../app/components/GitHubBackground";
import { ProjectsOverview } from "../../app/components/ProjectsOverview";
import { ExtensionEventLogger } from "../../app/tracking/ExtensionEventLogger";
import { eventLogger } from "../../app/utils/context";
import * as github from "../../app/utils/github";
import { Domain, getDomain, getGitHubRoute, parseURL } from "../../app/utils/index";

function ejectComponent(mount: HTMLElement): void {
	try {
		unmountComponentAtNode(mount);
		mount.remove();
	} catch (e) {
		console.error(e);
	}
}

export function injectGitHubApplication(): void {
	window.addEventListener("load", () => {
		injectModules();
		chrome.runtime.sendMessage({ type: "getIdentity" }, (identity) => {
			if (identity) {
				(eventLogger as ExtensionEventLogger).updatePropsForUser(identity);
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
			(eventLogger as ExtensionEventLogger).updatePropsForUser(ev.detail);
			chrome.runtime.sendMessage({ type: "setIdentity", identity: ev.detail });
		} else {
			console.error("sourcegraph:identify missing details");
		}
	});
}

function injectModules(): void {
	chrome.runtime.sendMessage({ type: "getSessionToken" }, (token) => {
		if (token) {
			useAccessToken(token);
		}
		injectBackgroundApp(<GitHubBackground />);
		injectBlobAnnotators();
		injectSourcegraphInternalTools();
	});
};

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

function injectBlobAnnotators(): void {
	const { repoURI, path, isDelta } = parseURL(window.location);
	if (!repoURI) {
		console.error("cannot determine repo URI");
		return;
	}

	const uri = repoURI;
	function addBlobAnnotator(file: HTMLElement, mount: HTMLElement): void {
		const {headFilePath, baseFilePath} = isDelta ? github.getDeltaFileName(file) : { headFilePath: path, baseFilePath: null };
		if (!headFilePath) {
			console.error("cannot determine file path");
			return;
		}

		if (file.className.includes("sg-blob-annotated")) {
			// make this function idempotent
			return;
		}
		file.className = `${file.className} sg-blob-annotated`;
		render(<BlobAnnotator headPath={headFilePath} repoURI={uri} fileElement={file} basePath={baseFilePath} />, mount);
	}

	const files = github.getFileContainers();
	for (const file of Array.from(files)) {
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
		// Remove class name; allows re-applying annotations.
		annotated.item(idx).className = annotated.item(idx).className.replace("sg-blob-annotated", "");
	}
}

function injectSourcegraphInternalTools(): void {
	if (document.getElementById("sourcegraph-projet-overview")) {
		return;
	}

	if (window.location.href === "https://github.com/orgs/sourcegraph/projects") {
		const container = document.querySelector("#projects-results")!.parentElement!.children[0];
		let mount = document.createElement("span");
		mount.id = "sourcegraph-projet-overview";
		(container as Element).insertBefore(mount, (container as Element).firstChild);
		render(<ProjectsOverview />, mount);
	}
}
