import * as React from "react";
import { render } from "react-dom";
import { BitbucketBlobAnnotator } from "../../app/components/BitbucketBlobAnnotator";
import { injectBackgroundApp } from "../../app/utils/injectBackgroundApp";
import { BitbucketBrowseUrl, BitbucketMode, BitbucketUrl } from "../../app/utils/types";
import * as bitbucket from "../../app/utils/bitbucket";

export function injectBitbucketApplication(): void {
	window.addEventListener("load", () => {
		injectModules();
	});
}

function injectModules(): void {
	injectBackgroundApp(null);
	injectBitbucketBlobAnnotators();
}

function injectBitbucketBlobAnnotators(): void {
	const bitbucketURL = bitbucket.getBitbucketState(global.window.location);
	if (!bitbucketURL) {
		return;
	}
	if (bitbucketURL.mode === BitbucketMode.Browse) {
		const browseUrl: BitbucketBrowseUrl = bitbucketURL as BitbucketBrowseUrl;
		const fileContent = bitbucket.getCodeBrowser();
		if (!fileContent) {
			return;
		}
		if (fileContent.classList.contains("sg-blob-annotated")) {
			return;
		}
		fileContent.classList.add("sg-blob-annotated");
		const mount = createBlobAnnotatorMount(fileContent, ".file-toolbar");
		render(<BitbucketBlobAnnotator path={browseUrl.path} repo={browseUrl.repo} projectCode={browseUrl.projectCode} blobElement={fileContent} rev={browseUrl.rev} />, mount);
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
