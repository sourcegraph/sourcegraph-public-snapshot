// tslint:disable: typedef ordered-imports

import * as React from "react";

import * as BuildActions from "sourcegraph/build/BuildActions";
import {taskClass, elapsed} from "sourcegraph/build/Build";
import {Component} from "sourcegraph/Component";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as classNames from "classnames";

import {Collapsible} from "sourcegraph/components/index";

import * as styles from "./styles/Build.css";

const updateLogIntervalMsec = 1500;

interface Props {
	task: any;
	logs: any;
}

type State = any;

export class Step extends Component<Props, State> {
	_updateLogIntervalID: any;

	constructor(props: Props) {
		super(props);
		this.state = {
			task: null,
			log: null,
		};
		this._updateLogIntervalID = null;
	}

	componentWillUnmount() {
		this._stopUpdateLog();
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

	reconcileState(state: State, props: Props) {
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

	onStateTransition(prevState: State, nextState: State) {
		if (prevState.task !== nextState.task) {
			Dispatcher.Backends.dispatch(new BuildActions.WantLog(nextState.task.Build.Repo, nextState.task.Build.ID, nextState.task.ID));
		}
	}

	render(): JSX.Element | null {
		return (
			<Collapsible collapsed={true}>
				<div className={classNames(styles.step_title, taskClass(this.state.task))}>
					{this.state.task.Label}
					<span style={{float: "right"}}>{elapsed(this.state.task)}</span>
				</div>
				<div className={styles.step_body}>
					{this.state.log && <pre>{this.state.log.log}</pre>}
				</div>
			</Collapsible>
		);
	}
}
