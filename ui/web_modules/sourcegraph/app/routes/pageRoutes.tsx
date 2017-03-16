import { PlainRoute } from "react-router";

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
import { ZapBetaPage } from "sourcegraph/page/ZapBetaPage";

const pages = {
	[rel.about]: AboutPage,
	[rel.plan]: MasterPlanPage,
	[rel.beta]: BetaPage,
	[rel.contact]: ContactPage,
	[rel.security]: SecurityPage,
	[rel.pricing]: PricingPage,
	[rel.terms]: TermsPage,
	[rel.privacy]: PrivacyPage,
	[rel.docs]: DocsPage,
	[rel.twittercasestudy]: TwitterCaseStudyPage,
	[rel.zapbeta]: ZapBetaPage
};

export const pageRoutes: PlainRoute[] = Object.keys(pages).map(key => ({
	path: key,
	getComponents: (location, callback) => {
		callback(null, {
			main: pages[key],
			footer: Footer,
		});
	},
}));
