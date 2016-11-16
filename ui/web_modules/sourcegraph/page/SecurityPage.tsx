// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Hero, Heading} from "sourcegraph/components";
import * as styles from "sourcegraph/page/Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import {PageTitle} from "sourcegraph/components/PageTitle";

export function SecurityPage(props: {}, {}) {
	return (
		<div>
			<PageTitle title="Security" />
			<Hero pattern="objects" color="dark" className={base.pv1}>
				<div className={styles.container}>
					<Heading level={3} color="white">Security</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<p className={styles.p}>Security is core to everything we do.</p>

				<p className={styles.p}>Beyond keeping your code safe, Sourcegraph will help your team <em>improve</em> the security of your own code, by disseminating knowledge and best practices about safe code usage to all of your developers.</p>

				<p className={styles.p}>If you have specific questions or concerns, contact us at <a href="mailto:security@sourcegraph.com">security@sourcegraph.com</a>.</p>

				<Heading level={4} underline="blue" className={styles.h}>Access Controls</Heading>

				<p className={styles.p}>Access to all internal systems is protected by multi-factor authentication. All application and user access logs are stored centrally and monitored. No employee can access user information unless explicitly authorized by the user.</p>

				<Heading level={4} underline="blue" className={styles.h}>Infrastructure</Heading>

				<ul>
					<li>All production systems are hosted on <a href="https://cloud.google.com/security/">Google Cloud Platform</a>.</li>
					<li>Multi-factor authentication is required on all accounts that have access to production systems.</li>
					<li>Authentication is handled using <a href="https://auth0.com/security">Auth0</a>.</li>
					<li>All external network communication between production services occur over TLS/SSH.</li>
					<li>External access to production systems is restricted by firewall to restricted IP ranges.</li>
				</ul>

				<Heading level={4} underline="blue" className={styles.h}>Site security</Heading>

				<ul>
					<li>All data from <a href="https://sourcegraph.com/">Sourcegraph.com</a> is transmitted over HTTPS.</li>
					<li>Monitoring services alert our 24/7 support team of potential attacks.</li>
				</ul>

				<Heading level={4} underline="blue" className={styles.h}>Application Security</Heading>
				<ul>
					<li>No private code is accessed by Sourcegraph without explicit permission from an authorized user.</li>
					<li>All language analysis is static.</li>
					<li>We employ a strict review policy for changes to core security, authentication, and permissions logic.</li>
				</ul>

				<Heading level={4} underline="blue" className={styles.h}>Bug reports</Heading>
				<p className={styles.p}>If you think that you have found a security issue, please email us at security@sourcegraph.com. Please do not publicly disclose the issue until we’ve addressed it.</p>
				<p className={styles.p}>We provide monetary rewards, up to $4,000, for open bounties. This is determined based on the percentage of users impacted, the likelihood of encountering the vulnerability under normal use of the site, and the severity of potential service disruption or data leakage.</p>

			</div>
		</div>
	);
}
