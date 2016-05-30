// @flow

import React from "react";
import {Link} from "react-router";
import {Hero, Heading} from "sourcegraph/components";
import styles from "./Page.css";
import base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import Helmet from "react-helmet";

function AboutPage(props, {signedIn}): React$Element {
	return (
		<div>
			<Helmet title="About" />
			<Hero pattern="objects" color="blue" className={base.pv5}>
				<div styleName="container">
					<Heading level="3" color="white">Sourcegraph is how developers discover and understand code.</Heading>
					</div>
			</Hero>
			<div styleName="content">
				<p styleName="p">Remember the last time you were <em>in flow</em> while coding? You got a ton done and felt great.</p>
				<p styleName="p">But staying in flow is hard. While coding, you often need to dig through Q&amp;A forums, decipher documentation, trace through code, or interrupt a teammate. Even a 2-minute context switch breaks your flow.</p>
				<p styleName="p">Sourcegraph helps you find answers in seconds, not minutes.</p>
				<Heading level="4" underline="blue" className={styles.h5}>How it works</Heading>
				<p styleName="p">Sourcegraph is a fast, semantic code search and cross-reference engine. Search for any function, type, or package, and see how other developers use it, globally. It's free for public and private projects, and <Link to="/pricing">paid plans</Link> are available.</p>

				<p styleName="p">
					You can <a href="https://sourcegraph.com/sourcegraph/sourcegraph">see the code that powers Sourcegraph</a>, released publicly as <a href="https://fair.io/">Fair Source</a>.
				</p>

				{!signedIn && <div styleName="cta">
					<GitHubAuthButton color="purple" className={base.mr3}>
						<strong>Sign up with GitHub</strong>
					</GitHubAuthButton>
				</div>}

				<Heading level="4" underline="blue" className={styles.h5}>The future sooner</Heading>
				<p styleName="p">From lifesaving medicine to self-driving cars, the future’s most groundbreaking innovations will all rely on code, in one way or another. With so much software to write in the coming decades, we all need a better way to discover and understand code&mdash;one that will finally free us from re-doing work that’s been done 50,000 times before.</p>

				<p styleName="p">At Sourcegraph, we help developers <em>bring the future sooner</em>&mdash;by turning great ideas into great software more efficiently.</p>
			</div>
		</div>
	);
}

export default CSSModules(AboutPage, styles);
