import * as React from "react";
import {Panel} from "sourcegraph/components";
import * as styles from "sourcegraph/user/settings/styles/Settings.css";

interface Props {
	main: JSX.Element;
}

export function SettingsMain(props: Props): JSX.Element {
	return (
		<div className={styles.container}>
			<div className={styles.main}>
				<Panel className={styles.panel}>{props.main}</Panel>
			</div>
		</div>
	);
}
