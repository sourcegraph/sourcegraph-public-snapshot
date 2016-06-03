// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/header.css";
import Loader from "./Loader";

class Header extends React.Component {
	static propTypes = {
		title: React.PropTypes.string.isRequired,
		subtitle: React.PropTypes.string,
		loading: React.PropTypes.bool,
	};

	render() {
		return (
			<div styleName="container">
				<div styleName="cloning-title">{this.props.title}</div>
				<div styleName="cloning-subtext">{this.props.loading ? <Loader /> : this.props.subtitle}</div>
			</div>
		);
	}
}

export default CSSModules(Header, styles);
