export interface Workspace {
	uri: string;
	rev: string;
}

export interface Blob extends Workspace {
	path: string;
}

export interface BlobPosition extends Blob {
	line: number;
	char?: number;
}

export interface Signature {
	person: Person;
	date: string;
}

export interface Person {
	name: string;
	email: string;
	gravatarHash: string;
}

export interface Hunk {
	startLine: number;
	endLine: number;
	startByte: number;
	endByte: number;
	rev: string;
	author: Signature;
	message: string;
}

export interface ActiveRepoResults {
	active: string[];
	inactive: string[];
}

export interface TooltipData {
	loading?: boolean;
	title?: string;
	doc?: string;
	j2dUrl?: string;
}

export interface Reference {
	range: {
		start: {
			character: number;
			line: number;
		};
		end: {
			character: number;
			line: number;
		};
	};
	uri: string;
	repoURI: string;
}

export interface RepoRevSpec {
	repoURI: string;
	rev: string;
}

export interface CodeCell {
	cell: HTMLElement;
	eventHandler: HTMLElement;
	line: number;
	isAddition?: boolean; // for diff views
	isDeletion?: boolean; // for diff views
}

export interface ParsedURL {
	uri?: string;
	rev?: string;
}

// BlobURL is the URL format for blob pages.
export interface BlobURL extends ParsedURL {
	path?: string;
	line?: number;
	char?: number;
	modal?: string; // e.g. "references"
	modalMode?: string; // e.g. ["", "local", "external"]
}

export interface User {
	UID: string;
	Login: string;
	Name?: string;
	IsOrganization?: boolean;
	AvatarURL?: string;
	Location?: string;
	Company?: string;
	HomepageURL?: string;
	Disabled?: boolean;
	Admin?: boolean;
	Betas?: string[];
	Write?: boolean;
	RegisteredAt?: any;
}

export interface EmailAddr {
	Email?: string;
	Verified?: boolean;
	Primary?: boolean;
	Guessed?: boolean;
	Blacklisted?: boolean;
}

export interface EmailAddrList {
	EmailAddrs?: EmailAddr[];
}

export interface ExternalToken {
	uid?: string;
	host?: string;
	token?: string;
	scope?: string;
}
