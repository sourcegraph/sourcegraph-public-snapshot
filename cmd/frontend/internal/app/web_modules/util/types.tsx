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
	isDelta: boolean;
	isBase: boolean;
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
	path?: string;
}

export interface SourcegraphURL extends ParsedURL { }
