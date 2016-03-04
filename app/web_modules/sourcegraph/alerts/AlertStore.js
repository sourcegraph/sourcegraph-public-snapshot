import update from "react/lib/update";

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as AlertActions from "sourcegraph/alerts/AlertActions";

export class AlertStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this._removeAlert = this._removeAlert.bind(this);
	}

	reset() {
		this.alerts = deepFreeze([]);
		this.counter = 0; // for assigning unique ids to alerts
	}

	_removeAlert(id) {
		this.alerts.forEach((alert, i) => {
			if (alert.id === id) {
				this.alerts = update(this.alerts, {$splice: [[i, 1]]});
			}
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case AlertActions.AddAlert:
			{
				const id = ++this.counter;
				this.alerts = update(this.alerts, {$push: [{html: action.html, id: id}]});
				if (action.autoDismiss) {
					setTimeout(() => {
						Dispatcher.dispatch(new AlertActions.RemoveAlert(id));
					}, 3000);
				}
				break;
			}

		case AlertActions.RemoveAlert:
			this._removeAlert(action.id);
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new AlertStore(Dispatcher);
