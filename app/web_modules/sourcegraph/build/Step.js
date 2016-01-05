import React from "react";
import classNames from "classnames";

import * as BuildActions from "sourcegraph/build/BuildActions";
import {elapsed, panelClass, taskClass} from "sourcegraph/build/Build";
import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";

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
		let panelCls = classNames(panelClass(this.state.task), "step");

		let explicitlyInteracted = typeof this.state.expanded !== "undefined";
		let expanded = (explicitlyInteracted && this.state.expanded) || (!explicitlyInteracted && !this.state.task.Success);
		let bodyClass = classNames({
			"panel-collapse": true,
			"collapse": true,
			"in": expanded,
		});

		let headerID = `T${this.state.task.ID}`;
		let bodyID = `T${this.state.task.ID}-log-body`;

		return (
			<div className={panelCls}>
				<div className="panel-heading" role="tab" id={headerID}>
					<div className="pull-right">{elapsed(this.state.task)}</div>
					<h5 className="panel-title">
						<a role="button" data-toggle="collapse"
							onClick={() => this.setState({expanded: !expanded})}
							data-parent={`task-${this.state.task.ParentID}-subtasks`} href={bodyID}>
							<span className={taskClass(this.state.task).text}><i className={taskClass(this.state.task).icon}></i> {this.state.task.Label}</span>
						</a>
					</h5>
				</div>
				<div id={bodyID} className={bodyClass} role="tabpanel" aria-labelledby={headerID}>
					<div className="panel-body">
						{this.state.log ? <pre className="build-log">{this.state.log.log}</pre> : null}
					</div>
				</div>
			</div>
		);
	}
}

Step.propTypes = {
	task: React.PropTypes.object.isRequired,
	logs: React.PropTypes.object.isRequired,
};

export default Step;
