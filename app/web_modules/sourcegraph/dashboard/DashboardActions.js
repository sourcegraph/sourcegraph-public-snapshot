// For the backend.

export class WantCreateRepo {
	constructor(name) {
		this.name = name;
	}
}

export class RepoCreated {
	constructor(repos) {
		this.repos = repos;
	}
}

export class WantAddMirrorRepos {
	constructor(repos) {
		this.repos = repos;
	}
}

export class MirrorReposAdded {
	constructor(mirrorData) {
		this.mirrorData = mirrorData;
	}
}

export class WantAddMirrorRepo {
	constructor(repo) {
		this.repo = repo;
	}
}

export class MirrorRepoAdded {
	constructor(repo, mirrorData) {
		this.repo = repo;
		this.mirrorData = mirrorData;
	}
}

export class WantInviteUser {
	constructor(email, permission) {
		this.email = email;
		this.permission = permission;
	}
}

export class UserInvited {
	constructor(user) {
		this.user = user;
	}
}

export class WantInviteUsers {
	constructor(emails) {
		this.emails = emails;
	}
}

export class UsersInvited {
	constructor(teammates) {
		this.teammates = teammates;
	}
}
