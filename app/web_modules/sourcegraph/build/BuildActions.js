export class WantBuild {
	constructor(repo, buildID, force) {
		this.repo = repo;
		this.buildID = buildID;
		this.force = force;
	}
}

export class WantBuilds {
	constructor(repo, search, force) {
		this.repo = repo;
		this.search = search;
		this.force = force;
	}
}

export class WantNewestBuildForCommit {
	constructor(repo, commitID, force) {
		this.repo = repo;
		this.commitID = commitID;
		this.force = force;
	}
}

export class BuildsFetchedForCommit {
	constructor(repo, commitID, builds) {
		this.repo = repo;
		this.commitID = commitID;
		this.builds = builds;
	}
}

export class BuildFetched {
	constructor(repo, buildID, build) {
		this.repo = repo;
		this.buildID = buildID;
		this.build = build;
	}
}

export class BuildsFetched {
	constructor(repo, builds, search) {
		this.repo = repo;
		this.builds = builds;
		this.search = search;
	}
}

export class CreateBuild {
	constructor(repo, commitID, branch, tag) {
		this.repo = repo;
		this.commitID = commitID;
		this.branch = branch;
		this.tag = tag;
	}
}

export class WantLog {
	constructor(repo, buildID, taskID) {
		this.repo = repo;
		this.buildID = buildID;
		this.taskID = taskID;
	}
}

export class LogFetched {
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
	constructor(repo, buildID, force) {
		this.repo = repo;
		this.buildID = buildID;
		this.force = force;
	}
}

export class TasksFetched {
	constructor(repo, buildID, tasks) {
		this.repo = repo;
		this.buildID = buildID;
		this.tasks = tasks;
	}
}
