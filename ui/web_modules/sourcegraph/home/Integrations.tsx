import * as React from "react";

import { context } from "sourcegraph/app/context";
import * as styles from "sourcegraph/home/styles/Integrations.css";
import { Events, LoggableEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface Props {
	location: any;
}

function Tool({ name, img, url, event }: { name: string, img: string, url: string, event: LoggableEvent }): JSX.Element {
	return <a href={url} target="_blank" className={styles.tool} onClick={() => { if (event) { event.logEvent(); } }}>
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
			<div>
				<div className={styles.tool_list}>
					<Tool
						name="Chrome"
						img="/img/Dashboard/google-chrome.svg"
						url="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack"
						event={Events.ToolsModalDownloadCTA_Clicked}
					/>
				</div>
			</div>
		);
	}
}
