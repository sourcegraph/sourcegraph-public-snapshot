// tslint:disable

import {Route} from "react-router";
import {rel, abs} from "sourcegraph/app/routePatterns";
import * as invariant from "invariant";

// isPage returns whether the location path refers to one of these
// static pages.
//
// NOTE: All static pages should be added to the OR-expression.
export function isPage(pathname: string): boolean {
	invariant(pathname, "no pathname");
	pathname = pathname.slice(1); // trim leading "/"
	return pathname === abs.about || pathname === abs.contact || pathname === abs.security ||
		pathname === abs.pricing || pathname === abs.terms || pathname === abs.privacy;
}

export const routes: any[] = [
	{
		path: rel.about,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/AboutPage").default,
				});
			});
		},
	},
	{
		path: rel.beta,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/BetaPage").default,
				});
			});
		},
	},
	{
		path: rel.contact,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/ContactPage").default,
				});
			});
		},
	},
	{
		path: rel.security,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/SecurityPage").default,
				});
			});
		},
	},
	{
		path: rel.pricing,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/PricingPage").default,
				});
			});
		},
	},
	{
		path: rel.terms,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/TermsPage").default,
				});
			});
		},
	},
	{
		path: rel.privacy,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					navContext: null,
					main: require("sourcegraph/page/PrivacyPage").default,
				});
			});
		},
	},
	{
		path: rel.browserExtFaqs,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/home/BrowserExtFaqsPage").default,
				});
			});
		},
	},
];
