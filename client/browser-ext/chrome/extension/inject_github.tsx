import * as React from "react";
import { render, unmountComponentAtNode } from "react-dom";
import { useAccessToken } from "../../app/backend/xhr";
import { BlobAnnotator } from "../../app/components/BlobAnnotator";
import { GitHubBackground } from "../../app/components/GitHubBackground";
import { ProjectsOverview } from "../../app/components/ProjectsOverview";
import { ExtensionEventLogger } from "../../app/tracking/ExtensionEventLogger";
import { eventLogger } from "../../app/utils/context";
import * as github from "../../app/utils/github";
import { getDomain, getGitHubRoute, parseURL } from "../../app/utils/index";
import * as tooltips from "../../app/utils/tooltips";
import { Domain, GitHubBlobUrl, GitHubMode, GitHubUrl } from "../../app/utils/types";
import { injectCodeSearch } from "./inject_code_search";

export function injectGitHubApplication(marker: HTMLElement): void {
	window.addEventListener("load", () => {
		document.body.appendChild(marker);
		injectModules();
		chrome.runtime.sendMessage({ type: "getIdentity" }, (identity) => {
			if (identity) {
				(eventLogger as ExtensionEventLogger).updatePropsForUser(identity);
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
		// ejectModules();
		(eventLogger as ExtensionEventLogger).updateIdentity();

		// Remove all ".sg-annotated"; this allows tooltip event handlers to be re-registered.
		Array.from(document.querySelectorAll(".sg-annotated")).forEach((item: any) => {
			if (item && item.classList) {
				item.classList.remove("sg-annotated");
			}
		});
		tooltips.hideTooltip();
		injectModules();
	});
}

function injectModules(): void {
	chrome.runtime.sendMessage({ type: "getSessionToken" }, (token) => {
		if (token) {
			useAccessToken(token);
		}
		injectBlobAnnotators();
		injectSourcegraphInternalTools();
		injectCodeSearch();
	});
}

function injectBlobAnnotators(): void {
	const { repoURI, isDelta } = parseURL(window.location);
	let { path } = parseURL(window.location);
	const gitHubState: GitHubUrl | null = github.getGitHubState(window.location.href);
	// TODO(uforic): Eventually, use gitHubState for everything, but for now, only use it when the branch should have a
	// slash in it to fix that bug
	if (gitHubState && gitHubState.mode === GitHubMode.Blob && (gitHubState as GitHubBlobUrl).rev.indexOf("/") > 0) {
		// correct in case branch has slash in it
		path = (gitHubState as GitHubBlobUrl).path;
	}
	if (!repoURI) {
		console.error("cannot determine repo URI");
		return;
	}

	const uri = repoURI;
	function addBlobAnnotator(file: HTMLElement, mount: HTMLElement): void {
		const { headFilePath, baseFilePath } = isDelta ? github.getDeltaFileName(file) : { headFilePath: path, baseFilePath: null };
		if (!headFilePath) {
			console.error("cannot determine file path");
			return;
		}

		// if (file.className.includes("sg-blob-annotated")) {
		// 	// make this function idempotent
		// 	return;
		// }
		// file.className = `${file.className} sg-blob-annotated`;
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


function injectSourcegraphInternalTools(): void {
	if (document.getElementById("sourcegraph-projet-overview")) {
		return;
	}

	if (window.location.href === "https://github.com/orgs/sourcegraph/projects") {
		const container = document.querySelector("#projects-results")!.parentElement!.children[0];
		const mount = document.createElement("span");
		mount.id = "sourcegraph-projet-overview";
		(container as Element).insertBefore(mount, (container as Element).firstChild);
		render(<ProjectsOverview />, mount);
	}
}
