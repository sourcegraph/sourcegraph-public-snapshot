/**
 * PhabUrl represents the state contained in a Phabricator URL.
 * PhabDiffusionUrl is the page state for code reading. 
 * PhabDifferentialUrl is the page state for pull requests i.e. diffusion.
 */
export interface PhabUrl {
	mode: PhabricatorMode;
}

export interface PhabDiffusionUrl extends PhabUrl {
	repoURI: string;
	branch: string;
	path: string;
	rev: string;
}

export interface PhabDifferentialUrl extends PhabUrl {
	differentialId: string;
	baseBranch: string;
	baseRepoURI: string;
	headBranch: string;
	headRepoURI: string;
}

export interface PhabRevisionUrl extends PhabUrl {
	repoUri: string;
	parentRev: string;
	childRev: string;
}

export interface PhabChangeUrl extends PhabUrl {
	repoURI: string;
	branch: string;
	path: string;
	rev: string;
	prevRev: string;
}

export enum PhabricatorMode {
	Diffusion = 1,
	Differential,
	Revision,
	Change,
}

export enum Domain {
	GITHUB,
	SGDEV_PHABRICATOR,
	SOURCEGRAPH,
}

export interface CodeCell {
	cell: HTMLElement;
	line: number;
	isAddition?: boolean; // for diff views
	isDeletion?: boolean; // for diff views
}

export interface PhabricatorCodeCell extends CodeCell {
	isLeftColumnInSplit: boolean;
	isUnified: boolean;
}

export interface GitHubURLData {
	user?: string;
	repo?: string;
	repoURI?: string;
	rev?: string;
	path?: string;
	isDelta?: boolean;
	isPullRequest?: boolean;
	isCommit?: boolean;
}

export interface GitHubUrl {
	mode: GitHubMode;
}

export interface GitHubBlobUrl {
	mode: GitHubMode;
	owner: string;
	repo: string;
	revAndPath: string;
	lineNumber: string | undefined;
	rev: string;
	path: string;
}

export interface GitHubPullUrl {
	mode: GitHubMode;
	owner: string;
	repo: string;
	view: string;
	id: number;
}

export enum GitHubMode {
	Blob,
	Commit,
	PullRequest,
}
