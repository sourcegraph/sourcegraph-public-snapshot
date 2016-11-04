// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Hero, Heading} from "sourcegraph/components";
import * as styles from "sourcegraph/page/Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import Helmet from "react-helmet";

export function SecurityPage(props: {}, {}) {
	return (
		<div>
			<Helmet title="Security" />
			<Hero pattern="objects" color="dark" className={base.pv1}>
				<div className={styles.container}>
					<Heading level={3} color="white">Security</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<p className={styles.p}>Security is core to everything we do.</p>

				<p className={styles.p}>Beyond keeping your code safe, Sourcegraph will help your team <em>improve</em> the security of your own code, by disseminating knowledge and best practices about safe code usage to all of your developers.</p>

				<p className={styles.p}>If you have specific questions or concerns, contact us at <a href="mailto:security@sourcegraph.com">security@sourcegraph.com</a>.</p>

				<Heading level={3} underline="blue" className={styles.h5}>Security measures</Heading>

				<Heading level={4} className={styles.h}>Site security</Heading>

				<ul>
					<li>All data from Sourcegraph.com is transmitted strictly over HTTPS.</li>
					<li>Best-practice HTTP security headers are enforced (Strict-Transport-Security, Content-Type-Options, XSS-Protection, Frame-Options).</li>
					<li>User session cookies are validated using HMAC to prevent forgery.</li>
					<li>Hosted third-party content and code is escaped or sanitized before it is displayed to the user.</li>
					<li>Monitoring services alert our support team 24x7 of potential attacks.</li>
				</ul>

				<Heading level={4} className={styles.h}>Infrastructure</Heading>

				<ul>
					<li>All production systems are hosted on Google Cloud Platform. See <a href="https://cloud.google.com/security/">the Google Cloud Platform security policy</a> for more information.</li>
					<li>Network communication between production services occurs strictly over TLS/SSH.</li>
					<li>External network communication to production systems occurs strictly over TLS/SSH.</li>
					<li>External access to production systems is restricted by firewall to known IP ranges.</li>
					<li>Two-factor authentication is required on all Google Cloud Platform accounts that have access to production systems.</li>
				</ul>

				<Heading level={4} className={styles.h}>System security</Heading>

				<ul>
					<li>The latest reported security vulnerabilities are tracked and patches are applied as soon as possible.</li>
				</ul>

				<Heading level={4} className={styles.h}>Application security</Heading>

				<ul>
					<li>All language analysis is static.</li>
					<li>No private code is pulled onto Sourcegraph without explicit permission from an authorized user.</li>
				</ul>

				<p className={styles.p}>Sourcegraph maintains application security over time with a 3-tiered defense policy against the introduction of security vulnerabilities:</p>

				<ul>
					<li>Every new feature with security implications must include tests for potential security flaws.</li>
					<li>We have a strict review policy for changes to core security, authentication, and permissions logic.</li>
					<li>We employ static analysis safeguards to detect if any code path can access private user data while bypassing the necessary permissions checks.</li>
				</ul>

				<Heading level={4} className={styles.h}>Access Controls</Heading>

				<p className={styles.p}>Access to all internal systems is protected by multi-factor authentication. All application and user access logs are stored centrally and monitored.</p>

				<p className={styles.p}>No employee can access user information unless explicitly authorized by the user.</p>

				<Heading level={3} underline="blue" className={styles.h5}>Bug reports</Heading>
				<p className={styles.p}>If you think that you have found a security issue, please email us at security@sourcegraph.com. We take all reports seriously. We provide monetary rewards for open bounties.</p>



			</div>
		</div>
	);
}
