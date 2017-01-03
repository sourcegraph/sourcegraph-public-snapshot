import { rel } from "sourcegraph/app/routePatterns";
import { AboutPage } from "sourcegraph/page/AboutPage";
import { BetaPage } from "sourcegraph/page/BetaPage";
import { ContactPage } from "sourcegraph/page/ContactPage";
import { DocsPage } from "sourcegraph/page/DocsPage";
import { MasterPlanPage } from "sourcegraph/page/MasterPlanPage";
import { PricingPage } from "sourcegraph/page/PricingPage";
import { PrivacyPage } from "sourcegraph/page/PrivacyPage";
import { SecurityPage } from "sourcegraph/page/SecurityPage";
import { TermsPage } from "sourcegraph/page/TermsPage";
import { TwitterCaseStudyPage } from "sourcegraph/page/TwitterCaseStudyPage";

export const pageRoutes: any[] = [
	{
		path: rel.about,
		getComponents: (location, callback) => {
			callback(null, {
				main: AboutPage,
			});
		},
	},
	{
		path: rel.plan,
		getComponents: (location, callback) => {
			callback(null, {
				main: MasterPlanPage,
			});
		},
	},
	{
		path: rel.beta,
		getComponents: (location, callback) => {
			callback(null, {
				main: BetaPage,
			});
		},
	},
	{
		path: rel.contact,
		getComponents: (location, callback) => {
			callback(null, {
				main: ContactPage,
			});
		},
	},
	{
		path: rel.security,
		getComponents: (location, callback) => {
			callback(null, {
				main: SecurityPage,
			});
		},
	},
	{
		path: rel.pricing,
		getComponents: (location, callback) => {
			callback(null, {
				main: PricingPage,
			});
		},
	},
	{
		path: rel.terms,
		getComponents: (location, callback) => {
			callback(null, {
				main: TermsPage,
			});
		},
	},
	{
		path: rel.privacy,
		getComponents: (location, callback) => {
			callback(null, {
				main: PrivacyPage,
			});
		},
	},
	{
		path: rel.docs,
		getComponents: (location, callback) => {
			callback(null, {
				main: DocsPage,
			});
		},
	},
	{
		path: rel.twittercasestudy,
		getComponents: (location, callback) => {
			callback(null, {
				main: TwitterCaseStudyPage,
			});
		},
	},
];
