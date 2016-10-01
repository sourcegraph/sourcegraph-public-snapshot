import { TreeEntry } from "sourcegraph/api";

export type Action =
	WantFile |
	FileFetched |
	Toast |
	ClearToast;

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

export class Toast {
	msg: string;
	timeout: number;

	constructor(msg: string, timeout?: number) {
		this.msg = msg;
		this.timeout = timeout || 2500;
	}
}

export class ClearToast {}
