import * as React from "react";

import * as BuildActions from "sourcegraph/build/BuildActions";
import {taskClass, elapsed} from "sourcegraph/build/Build";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";

import {Collapsible} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/Build.css";

const updateLogIntervalMsec = 1500;

class Step extends Component {
	constructor(props) {
		super(props);
		this.state = {
			task: null,
			log: null,
		};
		this._updateLogIntervalID = null;
	}

	componentWillUnmount() {
		this._stopUpdateLog();
		if (super.componentWillUnmount) super.componentWillUnmount();
	}

	_startUpdateLog() {
		if (this._updateLogIntervalID === null && !global.it) { // skip when testing
			this._updateLogIntervalID = setInterval(this._updateLog.bind(this), updateLogIntervalMsec);
		}
	}

	_stopUpdateLog() {
		if (this._updateLogIntervalID !== null && !global.it) { // skip when testing
			clearInterval(this._updateLogIntervalID);
			this._updateLogIntervalID = null;
		}
	}

	_updateLog() {
		if (this.state && this.state.task !== null) {
			Dispatcher.Backends.dispatch(new BuildActions.WantLog(this.state.task.Build.Repo, this.state.task.Build.ID, this.state.task.ID));
			if (this.state.task.EndedAt) {
				this._stopUpdateLog();
			}
		}
	}

	reconcileState(state, props) {
		if (state.task !== props.task) {
			state.task = props.task;

			// Reset log if showing a different task.
			state.log = null;
			this._stopUpdateLog();

			if (state.task !== null) {
				this._startUpdateLog();
			}
		}

		// Keep the log up to date by refreshing it as new entries
		// are added.
		let log = props.logs.get(state.task.Build.Repo, state.task.Build.ID, state.task.ID);
		if (log !== null) {
			state.log = log;
		}
	}

	onStateTransition(prevState, nextState) {
		if (prevState.task !== nextState.task) {
			Dispatcher.Backends.dispatch(new BuildActions.WantLog(nextState.task.Build.Repo, nextState.task.Build.ID, nextState.task.ID));
		}
	}

	render() {
		return (
			<Collapsible collapsed={true}>
				<div styleName={`step-title ${taskClass(this.state.task)}`}>
					{this.state.task.Label}
					<span style={{float: "right"}}>{elapsed(this.state.task)}</span>
				</div>
				<div styleName="step-body">
					{this.state.log && <pre>{this.state.log.log}</pre>}
				</div>
			</Collapsible>
		);
	}
}

Step.propTypes = {
	task: React.PropTypes.object.isRequired,
	logs: React.PropTypes.object.isRequired,
};

export default CSSModules(Step, styles, {allowMultiple: true});
