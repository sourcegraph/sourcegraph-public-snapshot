import React from "react";
import classNames from "classnames";

import {elapsed, panelClass, taskClass} from "sourcegraph/build/Build";
import Component from "sourcegraph/Component";
import Step from "sourcegraph/build/Step";

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

		let panelCls = classNames(panelClass(task), "top-level-task");

		return (
			<div className={panelCls}>
				<div className="panel-heading" role="tab">
					<div className="pull-right">{elapsed(task)}</div>
					<h4 className="panel-title" id={`T${task.ID}`}>
						<i className={taskClass(task).icon}></i> {task.Label}
					</h4>
				</div>
				<div className="panel-body">
					<div className="panel-group steps"
						role="tablist" aria-multiselectable="true">
						{this.state.subtasks.map((subtask) => <Step key={subtask.ID} task={subtask} logs={this.state.logs} />)}
					</div>
				</div>
			</div>
		);
	}
}

TopLevelTask.propTypes = {
	task: React.PropTypes.object.isRequired,
	subtasks: React.PropTypes.array.isRequired,
	logs: React.PropTypes.object.isRequired,
};

export default TopLevelTask;
