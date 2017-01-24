import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
import { Button, Heading, Hero, Panel } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { PageTitle } from "sourcegraph/components/PageTitle";
import * as base from "sourcegraph/components/styles/_base.css";
import { Checkmark } from "sourcegraph/components/symbols/Primaries";
import * as styles from "sourcegraph/page/Page.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export function PricingPage({ location }: { location: RouterLocation }): JSX.Element {
	const privateScopeRegEx = /(^repo,)|(,repo$)|(,repo,)|(^repo$)/;
	const hasPrivateAccess = context.gitHubToken ? privateScopeRegEx.test(context.gitHubToken.scope) : false;

	let communityPlanButton: JSX.Element = <LocationStateToggleLink href="/join" modalName="join" location={location}
		onToggle={(v) => AnalyticsConstants.Events.PricingCTA_Clicked.logEvent({ plan: "community", page_name: AnalyticsConstants.PAGE_PRICING })}>
		<Button block={true} className={styles.plan_cta || ""} color="purple">Sign up</Button>
	</LocationStateToggleLink>;
	let teamPlanButton: JSX.Element = <LocationStateToggleLink href="/join" modalName="join" location={location}
		onToggle={(v) => AnalyticsConstants.Events.PricingCTA_Clicked.logEvent({ plan: "team", page_name: AnalyticsConstants.PAGE_PRICING })}>
		<Button block={true} className={styles.plan_cta || ""} color="green">Start 14 days free</Button>
	</LocationStateToggleLink>;
	let enterprisePlanButton: JSX.Element = <Link to="/contact"
		onClick={(v) => AnalyticsConstants.Events.PricingCTA_Clicked.logEvent({ plan: "enterprise", page_name: AnalyticsConstants.PAGE_PRICING })}>
		<Button block={true} className={styles.plan_cta || ""} color="blue">Contact us</Button>
	</Link>;

	if (context.user) {
		if (hasPrivateAccess) {
			teamPlanButton = <Button block={true} outline={true} color="green" className={styles.plan_cta_noop || ""}><Checkmark className={styles.icon} /> Your current plan</Button>;
			communityPlanButton = <div />;
		} else {
			communityPlanButton = <Button block={true} outline={true} color="purple" className={styles.plan_cta_noop || ""}><Checkmark className={styles.icon} /> Your current plan</Button>;
		}
	}

	return (
		<div>
			<PageTitle title="Pricing" />
			<Hero color="transparent" className={base.pv5 || ""}>
				<div className={styles.container_wide}>
					<Heading level={2}>Pricing</Heading>
					<Heading level={5} className={styles.subtitle || ""}>Ship better software faster.</Heading>
				</div>
			</Hero>
			<div className={styles.content_wide}>
				<div className={styles.plans}>
					<div className={styles.plan}>
						<div className={styles.plan_box}>
							<Panel color="purple" hover={false} className={styles.plan_panel || ""}>
								<Heading level={3} color="white" align="center">Community</Heading>
								<Heading level={1} color="white" align="center" style={{ height: 90, }}><span className={styles.currency}>$</span><span className={styles.amount}>0</span></Heading>
								<p>for individuals and teams with public code</p>
							</Panel>
							{communityPlanButton}
						</div>
						<div className={styles.details}>
							<Heading level={4} color="purple">
								Community includes<br />
							</Heading>
							<Heading level={5} color="purple" underline="purple">
								<span className={styles.details_cumulative}>features developers love:</span>
							</Heading>
							<ul className={styles.details_list}>
								<li>Code Intelligence for all public code</li>
								<li>Semantic search, browsing, and cross-references across unlimited GitHub repositories</li>
								<li>Automatic usage examples</li>
								<li>Web browser &amp; editor integrations</li>
							</ul>
						</div>
						<div className={styles.plan_footer}>
							{communityPlanButton}
						</div>
					</div>

					<div className={styles.plan}>
						<div className={styles.plan_box}>
							<Panel color="green" hover={false} className={styles.plan_panel || ""}>
								<Heading level={3} color="white" align="center">Team</Heading>
								<Heading level={1} color="white" align="center" style={{ height: 90, }}><span className={styles.currency}>$</span><span className={styles.amount}>25</span><span className={styles.amount_per}>/user/month</span></Heading>
								<p>for individuals and teams with private code</p>
							</Panel>
							{teamPlanButton}
						</div>
						<div className={styles.details}>
							<Heading level={4} color="green">
								Team includes<br />
							</Heading>
							<Heading level={5} color="green" underline="green">
								<span className={styles.details_cumulative}>features teams want:</span>
							</Heading>
							<ul className={styles.details_list}>
								<li>Everything in Community</li>
								<li>Code Intelligence for all private code</li>
								<li>Priority support</li>
							</ul>
						</div>
						<div className={styles.plan_footer}>
							{teamPlanButton}
						</div>
					</div>

					<div className={styles.plan}>
						<div className={styles.plan_box}>
							<Panel color="blue" hover={false} className={styles.plan_panel || ""}>
								<Heading level={3} color="white" align="center">Enterprise</Heading>
								<Heading level={1} color="white" align="center" style={{ height: 90, }}><span className={styles.amount_contact_us}>Contact us</span></Heading>
								<p>for large teams and enterprises</p>
							</Panel>
							{enterprisePlanButton}
						</div>
						<div className={styles.details}>
							<Heading level={4} color="blue">
								Enterprise includes<br />
							</Heading>
							<Heading level={5} color="blue" underline="blue">
								<span className={styles.details_cumulative}>features CTOs demand:</span>
							</Heading>
							<ul className={styles.details_list}>
								<li>Everything in Team</li>
								<li>Support for git repositories hosted on-premises or in the cloud</li>
								<li>Integration with tools like GitHub Enterprise and Phabricator</li>
								<li>Unlimited API integrations</li>
								<li>Dedicated Customer Success Manager</li>
							</ul>
						</div>
						<div className={styles.plan_footer}>
							{enterprisePlanButton}
						</div>
					</div>

				</div>
				<p className={styles.footer}>Plans are billed annually. Questions? <Link to="/contact">Contact us.</Link></p>
			</div>
		</div>
	);
}
