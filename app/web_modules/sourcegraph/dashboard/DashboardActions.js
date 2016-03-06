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
