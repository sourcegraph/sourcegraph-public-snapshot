import * as React from "react";
import { RouterLocation } from "sourcegraph/app/router";
import { Heading, Hero, PageTitle } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { BetaInterestForm } from "sourcegraph/home/BetaInterestForm";
import * as styles from "sourcegraph/page/Page.css";

interface BetaPageProps {
	location: RouterLocation;
}

export function BetaPage(props: BetaPageProps): JSX.Element {
	return (
		<div>
			<PageTitle title="Beta" />
			<Hero pattern="objects" color="blue" className={base.pv5}>
				<div className={styles.container}>
					<Heading level={2}>Sourcegraph beta program</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<p className={styles.p}>As a Sourcegraph beta participant, you'll get early access to future releases, including:</p>
				<ul>
					<li className={styles.p}>Support for more programming languages</li>
					<li className={styles.p}>More editor integrations</li>
					<li className={styles.p}>Browser extensions for Firefox, Safari, Internet Explorer, etc.</li>
				</ul>
				<p className={styles.p}>Here's how it works:</p>
				<ul>
					<li className={styles.p}>Fill out the form below to join. We'll be in touch when we have something ready for you.</li>
					<li className={styles.p}>Please don't write publicly about unreleased features.</li>
					<li className={styles.p}><a href="mailto:support@sourcegraph.com">Report bugs</a> that you encounter.</li>
					<li className={styles.p}>Share feedback with us and help shape the future of Sourcegraph.</li>
				</ul>
				<br />
				<Heading level={3} underline="blue">Register for beta access</Heading>

				<BetaInterestForm location={props.location} />
			</div>
		</div>
	);
}
