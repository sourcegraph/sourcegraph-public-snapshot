import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { LoginPage } from "sourcegraph/user/Login";
import { SignupPage } from "sourcegraph/user/Signup";

export const userRoutes: PlainRoute[] = [
	{
		path: rel.login,
		getComponents: (location, callback) => {
			callback(null, { main: LoginPage });
		},
	},
	{
		path: rel.signup,
		getComponents: (location, callback) => {
			callback(null, { main: SignupPage });
		},
	}
];
