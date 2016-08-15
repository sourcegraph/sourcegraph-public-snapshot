export type Action =
	WantCommit |
	CommitFetched |
	WantFileList |
	FileListFetched |
	WantSrclibDataVersion |
	FetchedSrclibDataVersion;

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

export class WantSrclibDataVersion {
	repo: string;
	commitID: string;
	path: string | null;
	force: boolean;

	constructor(repo: string, commitID: string, path?: string | null, force?: boolean) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path || null;
		this.force = force || false;
	}
}

export class FetchedSrclibDataVersion {
	repo: string;
	commitID: string;
	path: string | null;
	version: string | null;

	constructor(repo: string, commitID: string, path?: string | null, version?: string | null) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path || null;
		this.version = version || null;
	}
}
