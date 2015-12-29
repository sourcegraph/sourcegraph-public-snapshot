import React from "react";

import * as BuildActions from "sourcegraph/build/BuildActions";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";

const updateLogIntervalMsec = 1500;

class BuildTask extends Component {
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
		super.componentWillUnmount();
	}

	_startUpdateLog() {
		if (this._updateLogIntervalID === null) {
			this._updateLogIntervalID = setInterval(this._updateLog.bind(this), updateLogIntervalMsec);
		}
	}

	_stopUpdateLog() {
		if (this._updateLogIntervalID !== null) {
			clearInterval(this._updateLogIntervalID);
			this._updateLogIntervalID = null;
		}
	}

	_updateLog() {
		if (this.state && this.state.task !== null) {
			Dispatcher.asyncDispatch(new BuildActions.WantLog(this.state.task.Build.Repo.URI, this.state.task.Build.ID, this.state.task.ID));
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
		let log = props.logs.get(state.task.Build.Repo.URI, state.task.Build.ID, state.task.ID);
		if (log !== null) {
			state.log = log;
		}
	}

	onStateTransition(prevState, nextState) {
		if (prevState.task !== nextState.task) {
			Dispatcher.asyncDispatch(new BuildActions.WantLog(nextState.task.Build.Repo.URI, nextState.task.Build.ID, nextState.task.ID));
		}
	}

	render() {
		return (
			<div>
				{this.state.log ? <pre className="build-log">{this.state.log.log}</pre> : null}
			</div>
		);
	}
}

BuildTask.propTypes = {
	task: React.PropTypes.object.isRequired,
	logs: React.PropTypes.object.isRequired,
};

export default BuildTask;
