import {AuthInfo, EmailAddr, ExternalToken, Settings, User} from "sourcegraph/user/index";

export type Action = WantAuthInfo
	| FetchedAuthInfo
	| WantUser
	| FetchedUser
	| WantEmails
	| FetchedEmails
	| FetchedGitHubToken
	| UpdateSettings
	| SubmitSignup
	| SubmitLogin
	| SubmitLogout
	| SubmitForgotPassword
	| SubmitResetPassword
	| SignupCompleted
	| LoginCompleted
	| LogoutCompleted
	| ForgotPasswordCompleted
	| ResetPasswordCompleted;

export class WantAuthInfo {
	accessToken: string;

	constructor(accessToken: string) {
		this.accessToken = accessToken;
	}
}

export class FetchedAuthInfo {
	accessToken: string;
	authInfo: AuthInfo | null; // null if unauthenticated

	constructor(accessToken: string, authInfo: AuthInfo | null) {
		this.accessToken = accessToken;
		this.authInfo = authInfo;
	}
}

export class WantUser {
	uid: number;

	constructor(uid: number) {
		this.uid = uid;
	}
}

export class FetchedUser {
	uid: number;
	user: User;

	constructor(uid: number, user: User) {
		this.uid = uid;
		this.user = user;
	}
}

export class WantEmails {
	uid: number;

	constructor(uid: number) {
		this.uid = uid;
	}
}

export class FetchedEmails {
	uid: number;
	emails: EmailAddr[];

	constructor(uid: number, emails: EmailAddr[]) {
		this.uid = uid;
		this.emails = emails;
	}
}

// No WantGitHubToken because it is included in the AuthInfo response.

export class FetchedGitHubToken {
	uid: number;
	token: ExternalToken;

	constructor(uid: number, token: ExternalToken) {
		this.uid = uid;
		this.token = token;
	}
}

export class SubmitSignup {
	login: string;
	password: string;
	email: string;

	constructor(login: string, password: string, email: string) {
		this.login = login;
		this.password = password;
		this.email = email;
	}
}

export class SignupCompleted {
	email: string;
	resp: any;
	eventName: string;
	signupChannel: string;

	constructor(email: string, resp: any) {
		this.email = email;
		this.resp = resp;
		this.eventName = "SignupCompleted";
		this.signupChannel = "email";
	}
}

export class SubmitLogin {
	login: string;
	password: string;

	constructor(login: string, password: string) {
		this.login = login;
		this.password = password;
	}
}

export class LoginCompleted {
	resp: any;
	eventName: string;

	constructor(resp: any) {
		this.resp = resp;
		this.eventName = "LoginCompleted";
	}
}

export class SubmitLogout {}

export class LogoutCompleted {
	resp: any;
	eventName: string;

	constructor(resp: any) {
		this.resp = resp;
		this.eventName = "LogoutCompleted";
	}
}

export class SubmitForgotPassword {
	email: string;

	constructor(email: string) {
		this.email = email;
	}
}

export class ForgotPasswordCompleted {
	resp: any;
	eventName: string;

	constructor(resp: any) {
		this.resp = resp;
		this.eventName = "ForgotPasswordCompleted";
	}
}

export class SubmitResetPassword {
	password: string;
	confirmPassword: string;
	token: string;

	constructor(password: string, confirmPassword: string, token: string) {
		this.password = password;
		this.confirmPassword = confirmPassword;
		this.token = token;
	}
}

export class ResetPasswordCompleted {
	resp: any;
	eventName: string;

	constructor(resp: any) {
		this.resp = resp;
		this.eventName = "ResetPasswordCompleted";
	}
}

export class SubmitBetaSubscription {
	email: string;
	firstName: string;
	lastName: string;
	languages: string[];
	editors: string[];
	message: string;
	// eventName purposefully left out

	constructor(email: string, firstName: string, lastName: string, languages: string[], editors: string[], message: string) {
		this.email = email;
		this.firstName = firstName;
		this.lastName = lastName;
		this.languages = languages;
		this.editors = editors;
		this.message = message;
	}
}

export class BetaSubscriptionCompleted {
	resp: any;
	eventName: string;

	constructor(resp: any) {
		this.resp = resp;
		this.eventName = "BetaSubscriptionCompleted";
	}
}

export class UpdateSettings {
	settings: Settings;

	constructor(settings: Settings) {
		this.settings = settings;
	}
}
