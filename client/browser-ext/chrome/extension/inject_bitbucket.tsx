import * as React from "react";
import { render } from "react-dom";
import { BitbucketBlobAnnotator } from "../../app/components/BitbucketBlobAnnotator";
import * as bitbucket from "../../app/utils/bitbucket";
import { BitbucketBrowseUrl, BitbucketMode, BitbucketUrl } from "../../app/utils/types";

export function injectBitbucketApplication(): void {
	window.addEventListener("load", () => {
		injectModules();
	});
}

function injectModules(): void {
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

		// Note: this assumes the user has configured the Sourcegraph instance to map repo URLs of the form
		// `bitbucket/${project}/${repo}` to `http(s)://${bitbucket_server_host}/scm/${project}/${repo}.git`
		const repoId = `bitbucket/${browseUrl.projectCode}/${browseUrl.repo}`;

		render(<BitbucketBlobAnnotator path={browseUrl.path} repo={repoId} projectCode={browseUrl.projectCode} blobElement={fileContent} rev={browseUrl.rev} />, mount);
	}
}

function createBlobAnnotatorMount(fileContainer: HTMLElement, buttonClass: string): HTMLElement | null {
	const existingMount = fileContainer.querySelector(".sourcegraph-app-annotator");
	if (existingMount) {
		// Make this function idempotent; no need to create a mount twice.
		return existingMount as HTMLElement;
	}

	const mountEl = document.createElement("div");
	mountEl.style.display = "inline";
	mountEl.className = "sourcegraph-app-annotator";

	const actionLinks = fileContainer.querySelector(buttonClass);
	if (!actionLinks) {
		return null;
	}
	actionLinks.appendChild(mountEl);
	return mountEl;
}
