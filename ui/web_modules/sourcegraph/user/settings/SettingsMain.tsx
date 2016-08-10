// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "./styles/Settings.css";
import {Panel} from "sourcegraph/components/index";

type Props = {
	main: JSX.Element,
};

export function SettingsMain(props) {
	return (
		<div className={styles.container}>
			<div className={styles.main}>
				<Panel className={styles.panel}>{props.main}</Panel>
			</div>
		</div>
	);
}
