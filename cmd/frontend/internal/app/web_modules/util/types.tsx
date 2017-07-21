export interface TooltipData {
	loading?: boolean;
	title?: string;
	doc?: string;
	j2dUrl?: string;
}

export interface ReferencesData {
	loading?: boolean;
	references?: Reference[];
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
