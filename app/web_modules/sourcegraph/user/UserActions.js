export class SubmitSignup {
	constructor(login, password, email) {
		this.login = login;
		this.password = password;
		this.email = email;
	}
}

export class SignupCompleted {
	constructor(resp) {
		this.resp = resp;
	}
}

export class SubmitLogin {
	constructor(login, password) {
		this.login = login;
		this.password = password;
	}
}

export class LoginCompleted {
	constructor(resp) {
		this.resp = resp;
	}
}

export class SubmitLogout {
	constructor() {}
}

export class LogoutCompleted {
	constructor(resp) {
		this.resp = resp;
	}
}

export class SubmitForgotPassword {
	constructor(email) {
		this.email = email;
	}
}

export class ForgotPasswordCompleted {
	constructor(resp) {
		this.resp = resp;
	}
}

export class SubmitResetPassword {
	constructor(password, confirmPassword, token) {
		this.password = password;
		this.confirmPassword = confirmPassword;
		this.token = token;
	}
}

export class ResetPasswordCompleted {
	constructor(resp) {
		this.resp = resp;
	}
}
