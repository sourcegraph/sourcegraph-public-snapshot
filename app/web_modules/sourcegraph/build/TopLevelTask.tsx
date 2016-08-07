// tslint:disable

import * as React from "react";

import {elapsed} from "sourcegraph/build/Build";
import Component from "sourcegraph/Component";
import Step from "sourcegraph/build/Step";

import * as styles from "./styles/Build.css";

class TopLevelTask extends Component<any, any> {
	reconcileState(state, props) {
		if (state.task !== props.task) {
			state.task = props.task;
		}

		if (state.subtasks !== props.subtasks) {
			state.subtasks = props.subtasks;
		}

		if (state.logs !== props.logs) {
			state.logs = props.logs;
		}
	}

	render(): JSX.Element | null {
		let task = this.state.task;

		return (
			<div>
				<div className={styles.top_level_task_header}>
					<span className={styles.header_label}>{task.Label}</span>
					<span className={styles.elapsed_label}>{elapsed(task)}</span>
				</div>
				{this.state.subtasks.map((subtask) => <Step key={subtask.ID} task={subtask} logs={this.state.logs} />)}
			</div>
		);
	}
}

(TopLevelTask as any).propTypes = {
	task: React.PropTypes.object.isRequired,
	subtasks: React.PropTypes.array.isRequired,
	logs: React.PropTypes.object.isRequired,
};

export default TopLevelTask;
