// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Hero, Heading} from "sourcegraph/components";
import * as styles from "sourcegraph/page/Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import Helmet from "react-helmet";
import {context} from "sourcegraph/app/context";

export function MasterPlanPage(props: {}) {
	return (
		<div>
			<Helmet title="Sourcegraph Master Plan" />
			<Hero pattern="objects" color="blue" className={base.pv4}>
				<div className={styles.container}>
					<Heading level={2} color="white">Sourcegraph Master Plan</Heading>
					<p className={styles.p}>What we're building and why it matters</p>
				</div>
			</Hero>
			<div className={styles.content}>
				{!context.user && <div className={styles.cta}>
					<GitHubAuthButton color="purple" className={base.mr3}>
						<strong>Sign up with GitHub</strong>
					</GitHubAuthButton>
				</div>}
			</div>
		</div>
	);
}
