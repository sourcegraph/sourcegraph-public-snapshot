// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/header.css";

class Header extends React.Component {
	static propTypes = {
		title: React.PropTypes.string.isRequired,
		subtitle: React.PropTypes.string,
	};

	render() {
		return (
			<div styleName="container">
				<div styleName="cloning-title">{this.props.title}</div>
				<div styleName="cloning-subtext">{this.props.subtitle}</div>
			</div>
		);
	}
}

export default CSSModules(Header, styles);
