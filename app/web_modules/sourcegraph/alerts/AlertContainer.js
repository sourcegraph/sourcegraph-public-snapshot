import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import AlertStore from "sourcegraph/alerts/AlertStore";
import * as AlertActions from "sourcegraph/alerts/AlertActions";

class AlertContainer extends Container {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.alerts = AlertStore.alerts;
	}

	stores() { return [AlertStore]; }

	render() {
		return (
			<div className="alert-container">
				{this.state.alerts.map((alert, i) =>
					<div className="alert alert-info" key={i}>
						<div className="alert-content">
							<i className="sg-icon sg-icon-close btn-icon alert-dismiss"
								onClick={_ => Dispatcher.dispatch(new AlertActions.RemoveAlert(alert.id))}></i>
							<span dangerouslySetInnerHTML={{__html: alert.html}} />
						</div>
					</div>
				)}
			</div>
		);
	}
}

AlertContainer.propTypes = {
};

export default AlertContainer;
