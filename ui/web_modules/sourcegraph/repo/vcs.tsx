export interface Signature {
	Name: string;
	Email: string;
	Date: string;
};

export interface Commit {
	Author: Signature;
	Committer: Signature;
	Message: string;
	Parents?: string[];
};

export interface Branch {
	Name: string;
	Head: string;
	Commit?: Commit;
};

export interface Tag {
	Name: string;
	CommitID: string;
};

export function sortBranches(branches: Branch[]): Branch[] {
	if (!branches) { return branches; }
	return branches.sort((a, b) => {
		const ac = a.Commit;
		const bc = b.Commit;
		if (!ac || !bc) { return 0; }
		const as = ac.Committer || ac.Author;
		const bs = bc.Committer || bc.Author;
		if (as.Date < bs.Date) { return 1; }
		if (as.Date > bs.Date) { return -1; }
		return 0;
	});
}

export function sortTags(tags: Tag[]): Tag[] {
	if (!tags) { return tags; }
	return tags.sort((a, b) => {
		if (a.Name > b.Name) { return -1; }
		if (a.Name < b.Name) { return 1; }
		return 0;
	});
}
