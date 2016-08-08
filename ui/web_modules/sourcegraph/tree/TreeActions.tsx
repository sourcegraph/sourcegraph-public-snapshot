// tslint:disable

export class WantCommit {
	repo: any;
	rev: any;
	path: any;

	constructor(repo, rev, path) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
	}
}

export class CommitFetched {
	repo: any;
	rev: any;
	path: any;
	commit: any;

	constructor(repo, rev, path, commit) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.commit = commit;
	}
}

export class WantFileList {
	repo: any;
	commitID: any;

	constructor(repo, commitID) {
		this.repo = repo;
		this.commitID = commitID;
	}
}

export class FileListFetched {
	repo: any;
	commitID: any;
	fileList: any;

	constructor(repo, commitID, fileList) {
		this.repo = repo;
		this.commitID = commitID;
		this.fileList = fileList;
	}
}

export class WantSrclibDataVersion {
	repo: any;
	commitID: any;
	path: any;
	force: any;

	constructor(repo, commitID, path?, force?) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path || null;
		this.force = force || null;
	}
}

export class FetchedSrclibDataVersion {
	repo: any;
	commitID: any;
	path: any;
	version: any;
	
	constructor(repo, commitID, path?, versionOrNull?) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path || null;
		this.version = versionOrNull || null;
	}
}
