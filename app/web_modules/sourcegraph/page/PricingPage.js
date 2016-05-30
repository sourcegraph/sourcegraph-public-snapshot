// @flow

import React from "react";
import {Hero, Heading, Panel, Button} from "sourcegraph/components";
import styles from "./Page.css";
import {Link} from "react-router";
import base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {CheckIcon} from "sourcegraph/components/Icons";
import Helmet from "react-helmet";

function PricingPage(props, {signedIn, eventLogger}): React$Element {
	return (
		<div>
			<Helmet title="Pricing" />
			<Hero color="transparent" className={base.pv5 || ""}>
				<div styleName="container-wide">
					<Heading level="1">Pricing</Heading>
					<Heading level="4" className={styles.subtitle || ""}>Sourcegraph is free to use for public and private code.</Heading>
				</div>
			</Hero>
			<div styleName="content-wide">
				<div styleName="plans">
					<div styleName="plan">
						<div styleName="plan-box">
							<Panel color="purple" inverse={true} hover={false} className={styles["plan-panel"] || ""}>
								<Heading level="3" color="white" align="center">Free</Heading>
								<Heading level="1" color="white" align="center"><span styleName="currency">$</span><span styleName="amount">0</span></Heading>
								<p>For individuals and teams, for public and private code</p>
							</Panel>
							{!signedIn && <Link to="/join"
								onClick={(v) => v && eventLogger.logEvent("ClickPricingCTA", {plan: "free"})}>
								<Button block={true} className={styles["plan-cta"] || ""} color="purple">Sign up</Button>
							</Link>}
							{signedIn && <Button block={true} outline={true} color="purple" className={styles["plan-cta-noop"] || ""}><CheckIcon styleName="icon" /> Your current plan</Button>}
						</div>
						<div styleName="details">
							<Heading level="4" color="purple" underline="purple">
								Free includes<br/>
								<span styleName="details-cumulative">Features every dev loves:</span>
							</Heading>
							<ul styleName="details-list">
								<li>Semantic search, browsing, and cross-references across unlimited GitHub repositories</li>
								<li>Single branch, latest commit only for private projects</li>
								<li>All branches, all commits for public projects</li>
								<li>Automatic usage examples</li>
								<li>Web browser &amp; editor integrations</li>
							</ul>
						</div>
					</div>

					<div styleName="plan">
						<div styleName="plan-box">
							<Panel color="green" inverse={true} hover={false} className={styles["plan-panel"] || ""}>
								<Heading level="3" color="white" align="center">Standard</Heading>
								<Heading level="1" color="white" align="center"><span styleName="currency">$</span><span styleName="amount">50</span></Heading>
								<p>per&nbsp;active&nbsp;user per&nbsp;month, first&nbsp;15&nbsp;users&nbsp;free</p>
							</Panel>
							<Link to="/contact"
								onClick={(v) => v && eventLogger.logEvent("ClickPricingCTA", {plan: "standard"})}>
								<Button block={true} className={styles["plan-cta"] || ""} color="green">Contact us</Button>
							</Link>
						</div>
						<div styleName="details">
							<Heading level="4" color="green" underline="green">
								Standard includes<br/>
								<span styleName="details-cumulative">Everything in Free, and:</span>
							</Heading>
							<ul styleName="details-list">
								<li>Unlimited branches and commits for private projects</li>
								<li>Mandatory 2-factor authentication</li>
								<li>Priority support</li>
							</ul>
						</div>
					</div>

					<div styleName="plan">
						<div styleName="plan-box">
							<Panel color="blue" inverse={true} hover={false} className={styles["plan-panel"] || ""}>
								<Heading level="3" color="white" align="center">Enterprise</Heading>
								<Heading level="1" color="white" align="center"><span styleName="currency">$</span><span styleName="amount">100</span></Heading>
								<p>per&nbsp;active&nbsp;user per&nbsp;month, first&nbsp;15&nbsp;users&nbsp;free</p>
							</Panel>
							<Link to="/contact"
								onClick={(v) => v && eventLogger.logEvent("ClickPricingCTA", {plan: "free"})}>
								<Button block={true} className={styles["plan-cta"] || ""} color="blue">Contact us</Button>
							</Link>
						</div>
						<div styleName="details">
							<Heading level="4" color="blue" underline="blue">
								Enterprise includes<br/>
								<span styleName="details-cumulative">Everything in Standard, and:</span>
							</Heading>
							<ul styleName="details-list">
								<li>Unlimited API integrations</li>
								<li>99.99% guaranteed uptime SLA</li>
								<li>24/7 support with 5-hour response time</li>
								<li>Option to run Sourcegraph in your own network</li>
							</ul>
						</div>
					</div>
				</div>
				<p styleName="footer">Plans are billed annually. Special pricing is available for qualified non-profit organizations. Questions? <Link to="/contact">Contact us.</Link></p>
			</div>
		</div>
	);
}
PricingPage.contextTypes = {
	signedIn: React.PropTypes.bool,
	eventLogger: React.PropTypes.object.isRequired,
};

export default CSSModules(PricingPage, styles);
