import * as React from "react";

import {elapsed} from "sourcegraph/build/Build";
import Component from "sourcegraph/Component";
import Step from "sourcegraph/build/Step";

import CSSModules from "react-css-modules";
import styles from "./styles/Build.css";

class TopLevelTask extends Component {
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

	render() {
		let task = this.state.task;

		return (
			<div>
				<div styleName={`top_level_task_header`}>
					<span styleName="header_label">{task.Label}</span>
					<span styleName="elapsed_label">{elapsed(task)}</span>
				</div>
				{this.state.subtasks.map((subtask) => <Step key={subtask.ID} task={subtask} logs={this.state.logs} />)}
			</div>
		);
	}
}

TopLevelTask.propTypes = {
	task: React.PropTypes.object.isRequired,
	subtasks: React.PropTypes.array.isRequired,
	logs: React.PropTypes.object.isRequired,
};

export default CSSModules(TopLevelTask, styles, {allowMultiple: true});
