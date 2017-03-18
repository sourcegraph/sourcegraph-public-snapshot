import * as React from "react";
import { render } from "react-dom";
import { PhabDifferentialBlobAnnotator } from "../../app/components/PhabDifferentialBlobAnnotator";
import { PhabDiffusionBlobAnnotator } from "../../app/components/PhabDiffusionBlobAnnotator";
import { getFilepathFromFile, getPhabricatorState, tryGetBlobElement } from "../../app/utils/phabricator";
import { CodeCell, PhabChangeUrl, PhabDifferentialUrl, PhabDiffusionUrl, PhabRevisionUrl, PhabricatorMode, PhabUrl } from "../../app/utils/types";


/**
 * injectPhabricatorBlobAnnotators finds file blocks on the dom that sould be annotated, and adds blob annotators to them.
 */
export function injectPhabricatorBlobAnnotators(): void {
	const phabURL = getPhabricatorState(global.window.location);
	if (!phabURL) {
		return;
	}
	if (phabURL.mode === PhabricatorMode.Diffusion) {
		const file = document.getElementsByClassName("phui-main-column")[0] as HTMLElement;
		const phabDiffusionUrl = phabURL as PhabDiffusionUrl;
		const filePath = phabDiffusionUrl.path;
		const blob = tryGetBlobElement(file);
		if (!blob) {
			return;
		}
		if (file.className.includes("sg-blob-annotated")) {
			// make this function idempotent
			return;
		}
		file.className = `${file.className} sg-blob-annotated`;
		const mount = createBlobAnnotatorMount(file, ".phui-header-action-links");
		render(<PhabDiffusionBlobAnnotator branch={phabDiffusionUrl.branch} path={filePath} repoURI={phabDiffusionUrl.repoURI} blobElement={blob} rev={phabDiffusionUrl.rev} />, mount);
	} else if (phabURL.mode === PhabricatorMode.Differential || phabURL.mode === PhabricatorMode.Revision || phabURL.mode === PhabricatorMode.Change) {
		const files = document.getElementsByClassName("differential-changeset") as HTMLCollectionOf<HTMLElement>;
		for (const file of Array.from(files)) {
			if (file.className.includes("sg-blob-annotated")) {
				// make this function idempotent
				return;
			}
			file.className = `${file.className} sg-blob-annotated`;
			const mount = createBlobAnnotatorMount(file, ".differential-changeset-buttons");
			if (!mount) {
				continue;
			}
			const filePath = getFilepathFromFile(file);
			if (phabURL.mode === PhabricatorMode.Differential) {
				const phabDifferentialUrl = phabURL as PhabDifferentialUrl;
				render(<PhabDifferentialBlobAnnotator blobElement={file} path={filePath} headRepoURI={phabDifferentialUrl.headRepoURI} headBranch={phabDifferentialUrl.headBranch} baseRepoURI={phabDifferentialUrl.baseRepoURI} baseBranch={phabDifferentialUrl.baseBranch} />, mount);
			} else if (phabURL.mode === PhabricatorMode.Revision) {
				const phabRevisionUrl = phabURL as PhabRevisionUrl;
				render(<PhabDifferentialBlobAnnotator blobElement={file} path={filePath} headRepoURI={phabRevisionUrl.repoUri} headBranch={phabRevisionUrl.childRev} baseRepoURI={phabRevisionUrl.repoUri} baseBranch={phabRevisionUrl.parentRev} />, mount);
			} else if (phabURL.mode === PhabricatorMode.Change) {
				const phabChangeUrl = phabURL as PhabChangeUrl;
				render(<PhabDifferentialBlobAnnotator blobElement={file} path={filePath} headRepoURI={phabChangeUrl.repoURI} headBranch={phabChangeUrl.rev} baseRepoURI={phabChangeUrl.repoURI} baseBranch={phabChangeUrl.prevRev} />, mount);
			}
		}
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
