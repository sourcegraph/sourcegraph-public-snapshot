import * as React from "react";

import {context} from "sourcegraph/app/context";
import {Button, Heading} from "sourcegraph/components";
import * as styles from "sourcegraph/home/styles/Integrations.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	location: any;
}

function Tool({name, img, url, event}: {name: string, img: string, url: string, event: AnalyticsConstants.LoggableEvent}): JSX.Element {
	return <a href={url} target="_blank" className={styles.tool} onClick={() => {if (event) { event.logEvent(); }}}>
		<img className={styles.img} src={`${context.assetsRoot}${img}`}></img>
		<div className={styles.caption}>{name}</div>
	</a>;
}

export class Integrations extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		return (
			<div className={this.props.location.state && this.props.location.state.modal === "integrations" ? "" : styles.container}>
				<div className={styles.menu}>
					<Heading level={7} color="gray">Browser extensions</Heading>
					<div className={styles.tool_list}>
						<Tool
							name={"Chrome"}
							img={"/img/Dashboard/google-chrome.svg"}
							url={"https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack"}
							event={AnalyticsConstants.Events.ToolsModalDownloadCTA_Clicked}
						/>
					</div>
				</div>
				{this.props.location.query.onboarding &&
					<footer className={styles.footer}>
						<a className={styles.footer_link} href="/desktop/home">
							<Button color="green" className={styles.footer_btn}>Continue</Button>
						</a>
					</footer>
				}
			</div>
		);
	}
}
