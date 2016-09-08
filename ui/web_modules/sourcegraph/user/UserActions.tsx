import {Settings} from "sourcegraph/user";

export type Action =
	UpdateSettings |
	SubmitSignup |
	SubmitLogin |
	SubmitForgotPassword |
	SubmitResetPassword |
	SignupCompleted |
	LoginCompleted |
	ForgotPasswordCompleted |
	ResetPasswordCompleted;

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
