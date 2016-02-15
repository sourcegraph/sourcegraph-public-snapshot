// For the backend.

export class WantAddRepos {
	constructor(repos) {
		this.repos = repos;
	}
}

export class ReposAdded {
	constructor(repos) {
		this.repos = repos;
	}
}

export class WantAddUsers {
	constructor(users) {
		this.users = users;
	}
}

export class UsersAdded {
	constructor(users) {
		this.users = users;
	}
}

// For AddRepos + AddUsers widgets.

export class SelectRepoOrg {
	constructor(org) {
		this.org = org;
	}
}

export class SelectUserOrg {
	constructor(org) {
		this.org = org;
	}
}

export class SelectRepos {
	constructor(repos, selectAll) {
		this.repos = repos;
		this.selectAll = selectAll;
	}
}

export class SelectRepo {
	constructor(repoKey, select) {
		this.repoKey = repoKey;
		this.select = select;
	}
}

export class SelectUsers {
	constructor(users, selectAll) {
		this.users = users;
		this.selectAll = selectAll;
	}
}

export class SelectUser {
	constructor(userKey, select) {
		this.userKey = userKey;
		this.select = select;
	}
}
