// @flow

import React from "react";
import {Hero, Heading} from "sourcegraph/components";
import styles from "./Page.css";
import base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import Helmet from "react-helmet";
import BetaInterestForm from "sourcegraph/home/BetaInterestForm";

function BetaPage(props, {signedIn}): React$Element {
	return (
		<div>
			<Helmet title="Beta" />
			<Hero pattern="objects" className={base.pv5}>
				<div styleName="container">
					<Heading level="2" color="orange">Become a Sourcegraph beta tester!</Heading>
				</div>
			</Hero>
			<div styleName="content">
				<p styleName="p">At Sourcegraph, we're constantly finding new ways to keep you <em>in flow</em> while you code. If you&#39;re up for trying out some beta-quality software, apply today and help us bring the future sooner.</p>

				<Heading level="3" underline="blue" className={styles.h5}>What is beta access?</Heading>
				<p styleName="p">As a Sourcegraph beta tester, you'll get access to the latest and greatest (but potentially unstable) features that we're preparing for all Sourcegraph users. This includes:</p>
				<ul>
					<li><p styleName="p">More languages: JavaScript, Python, C#, etc., as we build and improve them.</p></li>
					<li><p styleName="p">Sourcegraph for your editor of choice.</p></li>
					<li><p styleName="p">Sourcegraph for Firefox, Safari, etc.</p></li>
				</ul>
				<br/>
				<Heading level="3" underline="blue" className={styles.h4}>Register for beta access</Heading>

				{!signedIn && <div styleName="cta">
					<p styleName="p">You must sign in to continue.</p>
					<GitHubAuthButton returnTo="/beta" color="purple" className={base.mr3}>
						<strong>Sign in with GitHub</strong>
					</GitHubAuthButton>
				</div>}

				{signedIn && <BetaInterestForm />}
			</div>
		</div>
	);
}
BetaPage.contextTypes = {
	signedIn: React.PropTypes.bool,
};

export default CSSModules(BetaPage, styles);
