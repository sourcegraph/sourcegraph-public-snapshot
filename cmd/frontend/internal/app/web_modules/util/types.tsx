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
