import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import {context} from "sourcegraph/app/context";
import {Heading, Hero} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/page/Page.css";

export function AboutPage(): JSX.Element {
	return (
		<div>
			<Helmet title="About" />
			<Hero pattern="objects" color="blue" className={base.pv5}>
				<div className={styles.container}>
					<Heading level={3} color="white">Sourcegraph is how developers discover and understand code.</Heading>
					</div>
			</Hero>
			<div className={styles.content}>
				<Heading level={4} underline="purple" className={styles.h5}>Our purpose: the future sooner</Heading>
				<p className={styles.p}>From lifesaving medicine to self-driving cars, the future’s most groundbreaking innovations will all rely on code, in one way or another. With so much software to write in the coming decades, we all need a better way to discover and understand code&mdash;one that will finally free us from re-doing work that’s been done 50,000 times before.</p>
				<p className={styles.p}>At Sourcegraph, we help developers <em>bring the future sooner</em>&mdash;by turning great ideas into great software more efficiently.</p>
				<p className={styles.p}>In the last 24 hours, you almost certainly used a product built by developers who use Sourcegraph. If you want to help bring it to every developer, <a href="/jobs">join our team</a>.</p>
				<br/>

				<Heading level={4} className={styles.h5}><Link to="/plan"><u>Sourcegraph Master Plan &raquo;</u></Link></Heading>
				<p className={styles.p}>What we're doing and why it matters</p>
				<br/>

				<Heading level={4} className={styles.h5}><Link to="/docs"><u>Sourcegraph Documentation &raquo;</u></Link></Heading>
				<p className={styles.p}>How to use Sourcegraph and browser/editor integrations</p>
				<br/><br/>

				{!context.user && <div className={styles.cta}>
					<GitHubAuthButton color="purple" className={base.mr3}>
						<strong>Sign up with GitHub</strong>
					</GitHubAuthButton>
				</div>}

			</div>
		</div>
	);
}
