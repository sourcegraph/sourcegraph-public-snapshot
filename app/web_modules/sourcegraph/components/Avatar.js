import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/avatar.css";

const PLACEHOLDER_IMAGE = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";

class Avatar extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return (
			<img styleName={this.state.size ? this.state.size : "small"} src={this.state.img || PLACEHOLDER_IMAGE} />
		);
	}
}

Avatar.propTypes = {
	img: React.PropTypes.string,
	size: React.PropTypes.string,
};

export default CSSModules(Avatar, styles);
