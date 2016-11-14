// getFileContainers returns the elements on the page which should be marked
// up with tooltips & links:
// - blob view: a single file
// - commit view: one or more file diffs
// - PR conversation view: snippets with inline comments
// - PR unified/split view: one or more file diffs
export function getFileContainers(): HTMLCollectionOf<HTMLElement> {
	return document.getElementsByClassName("file") as HTMLCollectionOf<HTMLElement>;
}

// getDeltaFileName returns the path of the file container
export function getDeltaFileName(container: HTMLElement): string {
	const info = container.querySelector(".file-info") as HTMLElement;
	if (!info) {
		throw new Error("no .file-info child on delta file container");
	}

	if (info.title) {
		// for PR conversation snippets
		return info.title;
	} else {
		const link = info.querySelector("a") as HTMLElement;
		if (!link) {
			throw new Error("no <a> child on delta file container .file-info");
		}
		return link.title;
	}
}
