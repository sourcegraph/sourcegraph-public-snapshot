// For the backend.

export class WantCreateRepo {
	constructor(name) {
		this.name = name;
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

export class WantInviteUser {
	constructor(email, permission) {
		this.email = email;
		this.permission = permission;
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

export class DismissUsersModal {
	constructor() {
	}
}

export class DismissReposModal {
	constructor() {
	}
}

export class OpenAddUsersModal {
	constructor() {
	}
}

export class OpenAddReposModal {
	constructor() {
	}
}
