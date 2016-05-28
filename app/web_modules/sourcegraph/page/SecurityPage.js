// @flow

import React from "react";
import {Hero, Heading} from "sourcegraph/components";
import styles from "./Page.css";
import base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";

function SecurityPage(props, {signedIn}): React$Element {
	return (
		<div>
			<Hero pattern="objects" color="dark" className={base.pv1}>
				<div styleName="container">
					<Heading level="3" color="white">Sourcegraph Security</Heading>
				</div>
			</Hero>
			<div styleName="content">
				<p styleName="p">We know the security of your code is extremely important. Our business depends on keeping your code safe. We take this commitment very seriously and work every day to continually earn your trust.</p>

				<p styleName="p">Our commitment to security has always been strong because Sourcegraph's employees have extensive experience in building secure systems. Prior to Sourcegraph, our founders and early employees built software for highly sensitive environments, including top-5 U.S. financial institutions. Our CEO, Quinn Slack, also performed graduate-level research in the Secure Computer Systems group of the <a href="https://cs.stanford.edu/" target="_blank">Stanford Computer Science Department</a>, helping build <a href="https://en.wikipedia.org/wiki/Tcpcrypt" target="_blank">tcpcrypt</a> and submitting security-related patches to cURL, OpenSSL, GnuTLS, NSS, and Chromium.</p>

				<p styleName="p">Beyond keeping your code safe, Sourcegraph can actually help your team <em>improve</em> the security of your own code, by disseminating knowledge and best practices about safe code usage to all of your developers.</p>

				<p styleName="p">The security policy below details the steps we take to keep your code safe. Contact us at <a href="mailto:security@sourcegraph.com">security@sourcegraph.com</a> with questions and feedback.</p>

                <Heading level="3" underline="blue" className={styles.h5}>Security measures</Heading>

				<Heading level="4" className={styles.h}>Site security</Heading>

                <ul>
                  <li>All data from Sourcegraph.com is transmitted strictly over HTTPS.</li>
                  <li>Best-practice HTTP security headers are enforced (Strict-Transport-Security, Content-Type-Options, XSS-Protection, Frame-Options).</li>
                  <li>User session cookies are validated using HMAC to prevent forgery.</li>
                  <li>Hosted third-party content and code is escaped or sanitized before it is displayed to the user.</li>
                  <li>Monitoring services alert our support team 24x7 of potential attacks.</li>
                </ul>

                <Heading level="4" className={styles.h}>Infrastructure</Heading>

                <ul>
                  <li>All production systems are hosted on Google Cloud Platform. See <a href="https://cloud.google.com/security/">the Google Cloud Platform security policy</a> for more information.</li>
                  <li>Network communication between production services occurs strictly over TLS/SSH.</li>
                  <li>External network communication to production systems occurs strictly over TLS/SSH.</li>
                  <li>External access to production systems is restricted by firewall to known IP ranges.</li>
                  <li>Two-factor authentication is required on all Google Cloud Platform accounts that have access to production systems.</li>
                </ul>

                <Heading level="4" className={styles.h}>System security</Heading>

                <ul>
                    <li>The latest reported security vulnerabilities are tracked and patches are applied as soon as possible.</li>
                </ul>

                <Heading level="4" className={styles.h}>Application security</Heading>

				<p>The code that powers this site is publicly available in the <a href="https://sourcegraph.com/sourcegraph/sourcegraph" target="_blank">sourcegraph/sourcegraph repository</a>.</p>

                <ul>
                    <li>All language analysis is static.</li>
                    <li>No private code is pulled onto Sourcegraph without explicit permission from an authorized user.</li>
                </ul>

                <p styleName="p">Sourcegraph maintains application security over time with a 3-tiered defense policy against the introduction of security vulnerabilities:</p>

                <ul>
                    <li>Every new feature with security implications must include tests for potential security flaws.</li>
                    <li>We have a strict review policy for changes to core security, authentication, and permissions logic.</li>
                    <li>We employ static analysis safeguards to detect if any code path can access private user data while bypassing the necessary permissions checks.</li>
                </ul>

                <Heading level="4" className={styles.h}>Employee access</Heading>

                <p styleName="p">All access to critical internal systems (e.g., VMs, cloud storage, email) is protected by 2-factor authentication.</p>

                <p styleName="p">No Sourcegraph employee ever accesses private code unless explicitly authorized by a customer for support reasons.</p>

                <p styleName="p">All employee computers and devices are password-protected with full-disk encryption.</p>

                <Heading id="disclose" level="3" underline="blue" className={styles.h5}>Bug bounty program</Heading>

                <p styleName="p">Our bug bounty program is similar to ones offered by Google, Facebook, Mozilla, and GitHub, and offers a way for security researchers and engineers to report vulnerabilities in a responsible fashion and receive cash compensation for their efforts.</p>

                <p styleName="p">If you’ve found a vulnerability, email <a href="mailto:security@sourcegraph.com">security@sourcegraph.com</a> to report it. Please refer to the rules below for more information.</p>

                <Heading level="4" className={styles.h}>Rules</Heading>

                <ul>
                    <li>Don’t publicly disclose a bug before it’s fixed.</li>
                    <li>Don’t attempt to access another user’s account or data, and don’t disrupt another user’s usage of the site.</li>
                    <li>Don’t perform any attack that affects the stability or integrity of our site or our data. Spam attacks and DoS attacks are not allowed.</li>
                    <li>Only technical vulnerabilities are in scope. Don’t attempt social engineering, phishing, trespassing, or physical attacks.</li>
                    <li>If in doubt, <a href="mailto:security@sourcegraph.com">email us</a>.</li>
                </ul>

                <p styleName="p">On our end, we:</p>

                <ul>
                    <li>Won’t take legal action against you if you follow the rules.</li>
                    <li>Will respond quickly to your submission.</li>
                    <li>Will update you on the work to fix the vulnerability you reported.</li>
                    <li>Will credit you on this page, if you’d like us to.</li>
                </ul>

                <Heading level="4" className={styles.h}>Open bounties</Heading>

                <p styleName="p">Rewards for open bounties range from $10 to $4,000 and are determined at our discretion based on a variety of factors, including but not limited to the percentage of users impacted, the likelihood of encountering the vulnerability under normal use of the site, and the severity of potential service disruption or data leakage.</p>

                <ul>
                    <li>Sourcegraph.com: Sourcegraph.com is our main website. It is built in Go and uses a variety of open-source libraries, including <a href="https://srclib.org">srclib</a>.</li>
                    <li>Sourcegraph API: The Sourcegraph API is used by other applications to programatically interact with Sourcegraph. The API is rooted at <code>https://sourcegraph.com/.api</code>.</li>
                </ul>

                <Heading level="4" className={styles.h}>Bounty hunters</Heading>

                <table>
                    <tbody>
                        <tr>
                            <td><a href="https://twitter.com/robin7907">Robin Puri (Deep inder Singh Puri)</a></td>
                            <td>300 points</td>
                        </tr>
                        <tr>
                            <td><a href="https://www.facebook.com/nithish.varghese">Nithish Varghese</a></td>
                            <td>200 points</td>
                        </tr>
                    </tbody>
                </table>

				<Heading level="3" underline="blue" className={styles.h5}>Contact us</Heading>

                <p styleName="p">
                    Questions, concerns, or comments about our security policy? Contact us at <a href="mailto:security@sourcegraph.com">security@sourcegraph.com</a>.
                </p>
			</div>
		</div>
	);
}

export default CSSModules(SecurityPage, styles);
