import * as utils from ".";
import { BitbucketBrowseUrl, BitbucketMode, BitbucketUrl } from "./types";

const BB_BROWSE_REGEX = /^(https?):\/\/([A-Z\d\.-]{2,})(\.([A-Z]{2,}))?(:\d{2,4})?\/projects\/([A-Za-z0-9]+)\/repos\/([A-Za-z0-9]+)\/browse\/(.*)/i;

export function getBitbucketState(location: Location): BitbucketBrowseUrl | BitbucketUrl | null {
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
		};
	}
	return null;
}

export function getCodeBrowser(): HTMLElement | null {
	return document.getElementById("file-content");
}
