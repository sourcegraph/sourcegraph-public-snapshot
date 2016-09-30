export type Action =
	WantCommit |
	CommitFetched |
	WantFileList |
	FileListFetched;

export class WantCommit {
	repo: string;
	rev: string;
	path: string;

	constructor(repo: string, rev: string, path: string) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
	}
}

export class CommitFetched {
	repo: string;
	rev: string;
	path: string;
	commit: string;

	constructor(repo: string, rev: string, path: string, commit: string) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.commit = commit;
	}
}

export class WantFileList {
	repo: string;
	commitID: string;

	constructor(repo: string, commitID: string) {
		this.repo = repo;
		this.commitID = commitID;
	}
}

export class FileListFetched {
	repo: string;
	commitID: string;
	fileList: any;

	constructor(repo: string, commitID: string, fileList: any) {
		this.repo = repo;
		this.commitID = commitID;
		this.fileList = fileList;
	}
}
