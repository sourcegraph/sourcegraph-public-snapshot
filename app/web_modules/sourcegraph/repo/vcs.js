// @flow

type Signature = {
	Name: string;
	Email: string;
	Date: string;
};

type Commit = {
	Author: Signature;
	Committer: Signature;
	Message: string;
	Parents?: string[];
};

type Branch = {
	Name: string;
	Head: string;
	Commit?: Commit;
};

type Tag = {
	Name: string;
	CommitID: string;
};

export function sortBranches(branches: Array<Branch>): Array<Branch> {
	if (!branches) return branches;
	return branches.sort((a, b) => {
		const ac = a.Commit;
		const bc = b.Commit;
		if (!ac || !bc) return 0;
		const as = ac.Committer || ac.Author;
		const bs = bc.Committer || bc.Author;
		if (as.Date < bs.Date) return 1;
		if (as.Date > bs.Date) return -1;
		return 0;
	});
}

export function sortTags(tags: Array<Tag>): Array<Tag> {
	if (!tags) return tags;
	return tags.sort((a, b) => {
		if (a.Name > b.Name) return -1;
		if (a.Name < b.Name) return 1;
		return 0;
	});
}
