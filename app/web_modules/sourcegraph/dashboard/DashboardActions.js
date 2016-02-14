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
	constructor(repoIndex, select) {
		this.repoIndex = repoIndex;
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
	constructor(userIndex, select) {
		this.userIndex = userIndex;
		this.select = select;
	}
}
