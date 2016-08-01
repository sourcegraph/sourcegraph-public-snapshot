import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Settings.css";
import {Panel} from "sourcegraph/components";

function SettingsMain(props) {
	return (
		<div styleName="container">
			<div styleName="main">
				<Panel styleName="panel">{props.main}</Panel>
			</div>
		</div>
	);
}

SettingsMain.propTypes = {
	main: React.PropTypes.element,
};

export default CSSModules(SettingsMain, styles);
