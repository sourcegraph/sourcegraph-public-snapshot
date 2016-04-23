// @flow

import {rel} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";
import context from "sourcegraph/app/context";

const _onEnterAuthedUserRedirect = function(nextState, replace) {
	if (context.currentUser) {
		replace("/");
	}
};

const login = {
	onEnter: _onEnterAuthedUserRedirect,
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/user/Login").default,
			});
		});
	},
};
const signup = {
	onEnter: _onEnterAuthedUserRedirect,
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/user/Signup").default,
			});
		});
	},
};
const forgot = {
	onEnter: _onEnterAuthedUserRedirect,
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/user/ForgotPassword").default,
			});
		});
	},
};
const reset = {
	onEnter: _onEnterAuthedUserRedirect,
	getComponents: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/user/ResetPassword").default,
			});
		});
	},
};

export const routes: Array<Route> = [
	{
		...login,
		path: rel.login,
	},
	{
		...signup,
		path: rel.signup,
	},
	{
		...forgot,
		path: rel.forgot,
	},
	{
		...reset,
		path: rel.reset,
	},
];
