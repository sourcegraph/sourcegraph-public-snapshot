// tslint:disable

export class WantBuild {
	repo: any;
	buildID: any;
	force: any;

	constructor(repo, buildID, force?) {
		this.repo = repo;
		this.buildID = buildID;
		this.force = force;
	}
}

export class WantBuilds {
	repo: any;
	search: any;
	force: any;

	constructor(repo, search, force?) {
		this.repo = repo;
		this.search = search;
		this.force = force;
	}
}

export class WantNewestBuildForCommit {
	repo: any;
	commitID: any;
	force: any;

	constructor(repo, commitID, force) {
		this.repo = repo;
		this.commitID = commitID;
		this.force = force;
	}
}

export class BuildsFetchedForCommit {
	repo: any;
	commitID: any;
	builds: any;

	constructor(repo, commitID, builds) {
		this.repo = repo;
		this.commitID = commitID;
		this.builds = builds;
	}
}

export class BuildFetched {
	repo: any;
	buildID: any;
	build: any;

	constructor(repo, buildID, build) {
		this.repo = repo;
		this.buildID = buildID;
		this.build = build;
	}
}

export class BuildsFetched {
	repo: any;
	builds: any;
	search: any;

	constructor(repo, builds, search) {
		this.repo = repo;
		this.builds = builds;
		this.search = search;
	}
}

export class CreateBuild {
	repo: any;
	commitID: any;
	branch: any;
	tag: any;
	redirectAfterCreation: any;

	constructor(repo, commitID, branch, tag, redirectAfterCreation?) {
		this.repo = repo;
		this.commitID = commitID;
		this.branch = branch;
		this.tag = tag;
		this.redirectAfterCreation = redirectAfterCreation; // HACK
	}
}

export class WantLog {
	repo: any;
	buildID: any;
	taskID: any;

	constructor(repo, buildID, taskID) {
		this.repo = repo;
		this.buildID = buildID;
		this.taskID = taskID;
	}
}

export class LogFetched {
	repo: any;
	buildID: any;
	taskID: any;
	minID: any;
	maxID: any;
	log: any;

	constructor(repo, buildID, taskID, minID, maxID, log) {
		this.repo = repo;
		this.buildID = buildID;
		this.taskID = taskID;
		this.minID = minID;
		this.maxID = maxID;
		this.log = log;
	}
}

export class WantTasks {
	repo: any;
	buildID: any;
	force: any;

	constructor(repo, buildID, force?) {
		this.repo = repo;
		this.buildID = buildID;
		this.force = force;
	}
}

export class TasksFetched {
	repo: any;
	buildID: any;
	tasks: any;
	
	constructor(repo, buildID, tasks) {
		this.repo = repo;
		this.buildID = buildID;
		this.tasks = tasks;
	}
}
