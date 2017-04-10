import { PlainRoute } from "react-router";

import { rel } from "sourcegraph/app/routePatterns";
import { LoginPage } from "sourcegraph/user/Login";
import { SignupPage } from "sourcegraph/user/Signup";
import { Workbench } from "sourcegraph/workbench/workbench";

export const userRoutes: PlainRoute[] = [
	{
		path: rel.login,
		getComponents: (location, callback) => {
			callback(null, { main: Workbench, injectedComponent: LoginPage });
		},
	},
	{
		path: rel.signup,
		getComponents: (location, callback) => {
			callback(null, { main: Workbench, injectedComponent: SignupPage });
		},
	}
];
