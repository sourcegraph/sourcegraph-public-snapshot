// @flow

import React from "react";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import styles from "./styles/Settings.css";
import {Panel, TabItem} from "sourcegraph/components";

function SettingsMain({main, location}: {main: React$Element, location: Location}) {
	return (
		<div styleName="container">
			<nav styleName="nav">
				<Link to="/settings/repos">
					<TabItem color="blue" active={location.pathname === "/settings/repos"} direction="vertical">Repositories</TabItem>
				</Link>
			</nav>
			<Panel styleName="panel">{main}</Panel>
		</div>
	);
}
export default CSSModules(SettingsMain, styles);
