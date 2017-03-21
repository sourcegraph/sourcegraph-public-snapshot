import * as React from "react";

import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
import { Button, Heading, Hero, Panel } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { PageTitle } from "sourcegraph/components/PageTitle";
import * as base from "sourcegraph/components/styles/_base.css";
import { Checkmark } from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";
import * as styles from "sourcegraph/page/Page.css";
import { Events, PAGE_PRICING } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";

export function PricingPage({ location }: { location: RouterLocation }): JSX.Element {
	const privateScopeRegEx = /(^repo,)|(,repo$)|(,repo,)|(^repo$)/;
	const hasPrivateAccess = context.gitHubToken ? privateScopeRegEx.test(context.gitHubToken.scope) : false;

	let communityPlanButton: JSX.Element = <LocationStateToggleLink href="/join" modalName="join" location={location}
		onToggle={(v) => Events.PricingCTA_Clicked.logEvent({ plan: "community", page_name: PAGE_PRICING })}>
		<div style={{ width: "200px", margin: "auto", }}><Button block={true} className={styles.plan_cta || ""} color="orange">Sign up for free</Button></div>
	</LocationStateToggleLink>;
	let personalPlanButton: JSX.Element = <LocationStateToggleLink href="/join" modalName="join" location={location}
		onToggle={(v) => Events.PricingCTA_Clicked.logEvent({ plan: "personal", page_name: PAGE_PRICING })}>
		<Button block={true} className={styles.plan_cta || ""} color="purple">Sign up for free</Button>
	</LocationStateToggleLink>;
	let organizationPlanButton: JSX.Element = <LocationStateToggleLink href="/join" modalName="join" location={location}
		onToggle={(v) => Events.PricingCTA_Clicked.logEvent({ plan: "organization", page_name: PAGE_PRICING })}>
		<Button block={true} className={styles.plan_cta || ""} color="green">Start 14 days free</Button>
	</LocationStateToggleLink>;
	let enterprisePlanButton: JSX.Element = <a href="mailto:sales@sourcegraph.com"
		onClick={(v) => Events.PricingCTA_Clicked.logEvent({ plan: "enterprise", page_name: PAGE_PRICING })}>
		<Button block={true} className={styles.plan_cta || ""} color="blue">Contact us</Button>
	</a>;

	if (context.user) {
		communityPlanButton = <div />;
		if (hasPrivateAccess) {
			personalPlanButton = <Button block={true} outline={true} color="purple" className={styles.plan_cta_noop || ""}><Checkmark className={styles.icon} /> Full access during trial</Button>;
			organizationPlanButton = <Button block={true} outline={true} color="green" className={styles.plan_cta_noop || ""}><Checkmark className={styles.icon} /> Full access during trial</Button>;
		}
	}

	EventLogger.setUserViewedPricingPage();

	return (
		<div style={{ flex: 1 }}>
			<PageTitle title="Pricing" />
			<Hero color="transparent" className={base.pv5 || ""}>
				<div className={styles.container}>
					<Heading level={2}>Pricing</Heading>
					<p style={{ marginTop: whitespace[2], marginBottom: whitespace[3] }} className={styles.subtitle || ""}>Sourcegraph is <span style={{ backgroundColor: colors.greenL3(), color: colors.green(), fontWeight: "bold" }}>always free for public and open-source code</span>. Start using it for private code with a paid plan.</p>
					{communityPlanButton}
				</div>
			</Hero>
			<div className={styles.content_wide}>
				<div className={styles.plans}>
					<div className={styles.plan}>
						<div className={styles.plan_box}>
							<Panel color="purple" hover={false} className={styles.plan_panel || ""}>
								<Heading level={3} color="white" align="center">Personal</Heading>
								<Heading level={1} color="white" align="center" style={{}}><span className={styles.currency}>$</span><span className={styles.amount}>free</span></Heading>
								<span className={styles.amount_per}>free until Feb 1, 2018</span>
							</Panel>
							{personalPlanButton}
						</div>
						<div className={styles.details}>
							<p>Use Sourcegraph with your own private repositories.</p>
						</div>
						<div className={styles.plan_footer}>
							{personalPlanButton}
						</div>
					</div>

					<div className={styles.plan}>
						<div className={styles.plan_box}>
							<Panel color="green" hover={false} className={styles.plan_panel || ""}>
								<Heading level={3} color="white" align="center">Organization</Heading>
								<Heading level={1} color="white" align="center" style={{}}><span className={styles.currency}>$</span><span className={styles.amount}>10</span></Heading>
								<span className={styles.amount_per}>per user, per month</span>
							</Panel>
							{organizationPlanButton}
						</div>
						<div className={styles.details}>
							<p>Use Sourcegraph with your organization&apos;s private repositories.</p>
							<ul className={styles.details_list}>
								<li>Team permissions and billing</li>
								<li>Priority support</li>
							</ul>
						</div>
						<div className={styles.plan_footer}>
							{organizationPlanButton}
						</div>
					</div>

					<div className={styles.plan}>
						<div className={styles.plan_box}>
							<Panel color="blue" hover={false} className={styles.plan_panel || ""}>
								<Heading level={3} color="white" align="center">Enterprise</Heading>
								<Heading level={1} color="white" align="center" style={{}}><span className={styles.currency}>$</span><span className={styles.amount}>50</span></Heading>
								<span className={styles.amount_per}>per user, per month</span>
							</Panel>
							{enterprisePlanButton}
						</div>
						<div className={styles.details}>
							<p>Use Sourcegraph with code hosted on your own servers.</p>
							<ul className={styles.details_list}>
								<li>Integrate with GitHub Enterprise, Phabricator, and other tools</li>
								<li>Global code search</li>
								<li>Dedicated Customer Success Manager</li>
							</ul>
						</div>
						<div className={styles.plan_footer}>
							{enterprisePlanButton}
						</div>
					</div>

				</div>
				<p className={styles.footer}>
					Plans are billed annually. Questions? <a href="mailto:sales@sourcegraph.com">Contact us.</a>
				</p>
			</div>
		</div>
	);
}
