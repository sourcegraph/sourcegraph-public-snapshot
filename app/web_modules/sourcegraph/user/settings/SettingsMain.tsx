// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/Settings.css";
import {Panel} from "sourcegraph/components/index";

function SettingsMain(props) {
	return (
		<div styleName="container">
			<div styleName="main">
				<Panel styleName="panel">{props.main}</Panel>
			</div>
		</div>
	);
}

(SettingsMain as any).propTypes = {
	main: React.PropTypes.element,
};

export default CSSModules(SettingsMain, styles);
