// @flow

import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Settings.css";
import {Panel} from "sourcegraph/components";

function SettingsMain({main, location}: {main: React$Element<any>, location: Location}) {
	return (
		<div styleName="container">
			<div styleName="main">
				<Panel styleName="panel">{main}</Panel>
			</div>
		</div>
	);
}
export default CSSModules(SettingsMain, styles);
