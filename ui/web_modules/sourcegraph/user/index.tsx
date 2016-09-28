import {PlainRoute} from "react-router";
import {User} from "sourcegraph/api";
import {rel} from "sourcegraph/app/routePatterns";
import {Login} from "sourcegraph/user/Login";
import {Signup} from "sourcegraph/user/Signup";

// inBeta tells if the given user is a part of the given beta program.
export function inBeta(u: User | null, b: string): boolean {
	if (!u || !u.Betas) { return false; }
	return u.Betas.indexOf(b) !== -1;
}

export interface ExternalToken {
	uid: number;
	host: string;
	scope: string;
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
];
