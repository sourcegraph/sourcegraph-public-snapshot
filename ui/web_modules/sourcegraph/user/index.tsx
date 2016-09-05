import {PlainRoute} from "react-router";
import {User} from "sourcegraph/api";
import {rel} from "sourcegraph/app/routePatterns";
import {SearchSettings} from "sourcegraph/search/index";
import {ForgotPassword} from "sourcegraph/user/ForgotPassword";
import {Login} from "sourcegraph/user/Login";
import {ResetPassword} from "sourcegraph/user/ResetPassword";
import {Signup} from "sourcegraph/user/Signup";

// inBeta tells if the given user is a part of the given beta program.
export function inBeta(u: User | null, b: string): boolean {
	if (!u || !u.Betas) { return false; }
	return u.Betas.indexOf(b) !== -1;
}

// inAnyBeta tells if the given user is a part of ANY beta program.
export function inAnyBeta(u: User | null): boolean {
	if (!u) { return false; }
	return (u.Betas || []).length > 0;
}

// betaPending tells if the given user is registered for beta access but is not
// yet participating in any beta programs.
export function betaPending(u: User | null): boolean {
	if (!u) { return false; }
	return (u.BetaRegistered || false) && (u.Betas || []).length === 0;
}

export interface ExternalToken {
	uid: number;
	host: string;
	scope: string;
};

export interface Settings {
	search: SearchSettings | null;
};

export const routes: PlainRoute[] = [
	{
		path: rel.login,
		getComponents: (location, callback) => {
			callback(null, {
				main: Login,
			});
		},
	},
	{
		path: rel.signup,
		getComponents: (location, callback) => {
			callback(null, {
				main: Signup,
			});
		},
	},
	{
		path: rel.forgot,
		getComponents: (location, callback) => {
			callback(null, {
				main: ForgotPassword,
			});
		},
	},
	{
		path: rel.reset,
		getComponents: (location, callback) => {
			callback(null, {
				main: ResetPassword,
			});
		},
	},
];
