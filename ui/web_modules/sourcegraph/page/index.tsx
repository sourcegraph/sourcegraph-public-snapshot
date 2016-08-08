// tslint:disable

import {Route} from "react-router";
import {rel, abs} from "sourcegraph/app/routePatterns";
import * as invariant from "invariant";
import AboutPage from "sourcegraph/page/AboutPage";
import BetaPage from "sourcegraph/page/BetaPage";
import ContactPage from "sourcegraph/page/ContactPage";
import SecurityPage from "sourcegraph/page/SecurityPage";
import PricingPage from "sourcegraph/page/PricingPage";
import TermsPage from "sourcegraph/page/TermsPage";
import PrivacyPage from "sourcegraph/page/PrivacyPage";
import BrowserExtFaqsPage from "sourcegraph/home/BrowserExtFaqsPage";

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
			callback(null, {
				navContext: null,
				main: AboutPage,
			});
		},
	},
	{
		path: rel.beta,
		getComponents: (location, callback) => {
			callback(null, {
				navContext: null,
				main: BetaPage,
			});
		},
	},
	{
		path: rel.contact,
		getComponents: (location, callback) => {
			callback(null, {
				navContext: null,
				main: ContactPage,
			});
		},
	},
	{
		path: rel.security,
		getComponents: (location, callback) => {
			callback(null, {
				navContext: null,
				main: SecurityPage,
			});
		},
	},
	{
		path: rel.pricing,
		getComponents: (location, callback) => {
			callback(null, {
				navContext: null,
				main: PricingPage,
			});
		},
	},
	{
		path: rel.terms,
		getComponents: (location, callback) => {
			callback(null, {
				navContext: null,
				main: TermsPage,
			});
		},
	},
	{
		path: rel.privacy,
		getComponents: (location, callback) => {
			callback(null, {
				navContext: null,
				main: PrivacyPage,
			});
		},
	},
	{
		path: rel.browserExtFaqs,
		getComponents: (location, callback) => {
			callback(null, {
				main: BrowserExtFaqsPage,
			});
		},
	},
];
