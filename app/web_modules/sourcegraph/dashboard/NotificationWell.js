import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/NotificationWell.css";

class NotificationWell extends Component {
	constructor(props) {
		super(props);
	}
	reconcileState(state, props) {
		Object.assign(state, props);

	}
	render() {
		return (
			<div styleName={this.state.visible ? "notification-box" : "hidden"}>
				{this.state.children}
			</div>
		);
	}
}

NotificationWell.propTypes = {
	visible: React.PropTypes.bool.isRequired,
};

export default CSSModules(NotificationWell, styles);
