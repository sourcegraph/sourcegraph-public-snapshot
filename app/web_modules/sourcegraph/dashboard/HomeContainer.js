import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/Home.css";

class HomeContainer extends Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
	};

	constructor(props) {
		super(props);
	}

	render() {
		return <div styleName="home"> this </div>;
	}
}

export default CSSModules(HomeContainer, styles);
