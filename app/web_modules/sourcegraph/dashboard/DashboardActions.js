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
	constructor(emails) {
		this.emails = emails;
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
	constructor(repoURI, select) {
		this.repoURI = repoURI;
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
	constructor(login, select) {
		this.login = login;
		this.select = select;
	}
}
