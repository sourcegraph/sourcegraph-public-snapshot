export class WantBuild {
	constructor(repo, buildID, force) {
		this.repo = repo;
		this.buildID = buildID;
		this.force = force;
	}
}

export class BuildFetched {
	constructor(repo, buildID, build) {
		this.repo = repo;
		this.buildID = buildID;
		this.build = build;
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
