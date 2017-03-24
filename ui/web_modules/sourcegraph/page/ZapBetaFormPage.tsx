import * as React from "react";
import { RouterLocation } from "sourcegraph/app/router";
import { Heading, Hero } from "sourcegraph/components";
import { PageTitle } from "sourcegraph/components/PageTitle";
import * as base from "sourcegraph/components/styles/_base.css";
import { ZapBetaInterestForm } from "sourcegraph/home/ZapBetaInterestForm";
import * as styles from "sourcegraph/page/Page.css";

interface ZapBetaFormPageProps {
	location: RouterLocation;
}

export function ZapBetaFormPage(props: ZapBetaFormPageProps): JSX.Element {
	return (
		<div>
			<PageTitle title="Zap" />
			<Hero pattern="objects" className={base.pv5}>
				<div className={styles.container}>
					<Heading level={2} color="blue">Real-time code collaboration + intelligence</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<p className={styles.p}>Sourcegraph extends your editor to the web so you can share work-in-progress code instantly with teammates. You can also access global code intelligence with a simple hotkey from your editor.</p>
				<p className={styles.p}>Sound interesting? Sign up and we will update you when your editor is supported.</p>
				<ZapBetaInterestForm location={props.location} />
			</div>
		</div>
	);
}
