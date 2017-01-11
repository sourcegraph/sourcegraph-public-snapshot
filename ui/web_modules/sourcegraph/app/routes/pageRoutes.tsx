import { Footer } from "sourcegraph/app/Footer";
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
				footer: Footer,
			});
		},
	},
	{
		path: rel.plan,
		getComponents: (location, callback) => {
			callback(null, {
				main: MasterPlanPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.beta,
		getComponents: (location, callback) => {
			callback(null, {
				main: BetaPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.contact,
		getComponents: (location, callback) => {
			callback(null, {
				main: ContactPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.security,
		getComponents: (location, callback) => {
			callback(null, {
				main: SecurityPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.pricing,
		getComponents: (location, callback) => {
			callback(null, {
				main: PricingPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.terms,
		getComponents: (location, callback) => {
			callback(null, {
				main: TermsPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.privacy,
		getComponents: (location, callback) => {
			callback(null, {
				main: PrivacyPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.docs,
		getComponents: (location, callback) => {
			callback(null, {
				main: DocsPage,
				footer: Footer,
			});
		},
	},
	{
		path: rel.twittercasestudy,
		getComponents: (location, callback) => {
			callback(null, {
				main: TwitterCaseStudyPage,
				footer: Footer,
			});
		},
	},
];
