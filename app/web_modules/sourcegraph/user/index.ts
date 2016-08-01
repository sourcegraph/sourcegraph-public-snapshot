import {rel} from "sourcegraph/app/routePatterns";
import {SearchSettings} from "sourcegraph/search/index";

export interface User {
	UID: number;
	Login: string;
	Betas: string[];
	BetaRegistered: boolean;
};

// inBeta tells if the given user is a part of the given beta program.
export function inBeta(u: User | null, b: string): boolean {
	if (!u || !u.Betas) { return false; }
	return u.Betas.indexOf(b) !== -1;
}

// inAnyBeta tells if the given user is a part of ANY beta program.
export function inAnyBeta(u: User | null): boolean {
	if (!u) { return false; }
	return u.Betas.length > 0;
}

// betaPending tells if the given user is registered for beta access but is not
// yet participating in any beta programs.
export function betaPending(u: User | null): boolean {
	if (!u) { return false; }
	return u.BetaRegistered && u.Betas.length === 0;
}

export interface AuthInfo {
	UID?: number;
	Login?: string;
	Admin?: boolean;
};

export interface EmailAddr {
	Email: string;
};

export interface ExternalToken {
	uid: number;
	host: string;
	scope: string;
};

export interface Settings {
	search: SearchSettings | null;
};

export const routes: Array<ReactRouter.PlainRoute> = [
	{
		path: rel.login,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/user/Login").default,
				});
			});
		},
	},
	{
		path: rel.signup,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/user/Signup").default,
				});
			});
		},
	},
	{
		path: rel.forgot,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/user/ForgotPassword").default,
				});
			});
		},
	},
	{
		path: rel.reset,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/user/ResetPassword").default,
				});
			});
		},
	},
];
