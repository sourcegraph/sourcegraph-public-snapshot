import * as React from "react";
import { render } from "react-dom";
import { BitbucketBlobAnnotator } from "../../app/components/BitbucketBlobAnnotator";
import { injectBackgroundApp } from "../../app/utils/injectBackgroundApp";
import { BitbucketBrowseUrl, BitbucketMode, BitbucketUrl } from "../../app/utils/types";

export function injectBitbucketApplication(): void {
	window.addEventListener("load", () => {
		injectModules();
	});
}

function injectModules(): void {
	injectBackgroundApp(null);
	injectBitbucketBlobAnnotators();
}

const BB_BROWSE_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})(\.([A-Z]{2,}))?(:\d{2,4})?\/projects\/([A-Za-z0-9]+)\/repos\/([A-Za-z0-9]+)\/browse\/(.*)/i;

function getBitbucketState(location: Location): BitbucketUrl | null {
	const browseMatch = BB_BROWSE_REGEX.exec(location.href);
	if (browseMatch) {
		const match = {
			protocol: browseMatch[1],
			hostname: browseMatch[2],
			extension: browseMatch[4],
			port: browseMatch[5],
			projectCode: browseMatch[6],
			repo: browseMatch[7],
			path: browseMatch[8],
		};
		return {
			mode: BitbucketMode.Browse,
			projectCode: match.projectCode,
			repo: match.repo,
			path: match.path,
			rev: "master",
		} as BitbucketBrowseUrl;
	}
	return null;
}

function injectBitbucketBlobAnnotators(): void {
	const bitbucketURL = getBitbucketState(global.window.location);
	if (!bitbucketURL) {
		return;
	}
	if (bitbucketURL.mode === BitbucketMode.Browse) {
		const browseUrl: BitbucketBrowseUrl = bitbucketURL as BitbucketBrowseUrl;
		const fileContent = document.getElementById("file-content");
		if (!fileContent) {
			return;
		}
		if (fileContent.classList.contains("sg-blob-annotated")) {
			return;
		}
		fileContent.classList.add("sg-blob-annotated");
		const mount = createBlobAnnotatorMount(fileContent, ".file-toolbar");
		render(<BitbucketBlobAnnotator path={browseUrl.path} repo={"github.com/gorilla/mux"} projectCode={browseUrl.projectCode} blobElement={fileContent} rev={browseUrl.rev} />, mount);
	}
}

function createBlobAnnotatorMount(fileContainer: HTMLElement, buttonClass: string): HTMLElement | null {
	const existingMount = fileContainer.querySelector(".sourcegraph-app-annotator");
	if (existingMount) {
		// Make this function idempotent; no need to create a mount twice.
		return existingMount as HTMLElement;
	}

	const mountEl = document.createElement("div");
	mountEl.style.display = "inline-block";
	mountEl.className = "sourcegraph-app-annotator";

	const actionLinks = fileContainer.querySelector(buttonClass);
	if (!actionLinks) {
		return null;
	}
	actionLinks.appendChild(mountEl);
	return mountEl;
}
