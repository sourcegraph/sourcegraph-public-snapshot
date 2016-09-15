import { TreeEntry } from "sourcegraph/api";

export type Action =
	WantFile |
	FileFetched;

export class WantFile {
	repo: string;
	commitID: string;
	path: string;

	constructor(repo: string, commitID: string, path: string) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path;
	}
}

export class FileFetched {
	repo: string;
	commitID: string | null;
	path: string;
	file: TreeEntry;
	eventName: string;

	constructor(repo: string, commitID: string | null, path: string, file: TreeEntry) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path;
		this.file = file;
	}
}
