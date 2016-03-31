import React from "react";

import Component from "sourcegraph/Component";
import Styles from "./styles/NotificationWell.css";

class NotificationWell extends Component {
	constructor(props) {
		super(props);
		this.state = {isVisible: this.props.initVisible};
		this._handleClick = this._handleClick.bind(this);
	}
	reconcileState(state, props) {
		Object.assign(state, props);

	}
	_handleClick(event) {
		this.setState({isVisible: !this.state.isVisible});
	}
	render() {
		// TODO: Make this dismissable as an option
		// <div className={Styles.notification_box_x} onClick={this._handleClick}>X</div>
		return (
			<div className={this.state.isVisible? Styles.notification_box : Styles.notification_box_hidden}>
					{this.state.children}
			</div>
		);
	}
}

NotificationWell.propTypes = {
	initVisible: React.PropTypes.bool.isRequired,
};

export default NotificationWell;
