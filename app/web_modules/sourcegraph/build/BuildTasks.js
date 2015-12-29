import React from "react";
import URI from "urijs";

import BuildTask from "sourcegraph/build/BuildTask";
import Component from "sourcegraph/Component";

class BuildTasks extends Component {
	constructor(props) {
		super(props);
		this.state = {
			activeTaskID: null,
			activeTask: null,
		};
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		if (state.tasks !== props.tasks) {
			state.tasks = props.tasks;
		}

		// Update and respond to URL route.
		state.activeTaskID = this._activeTaskIDFromURL(URI.parse(props.location));

		if (state.tasks && state.tasks.length > 0 && state.activeTaskID === null) {
			state.activeTaskID = state.tasks[0].ID;
		}

		if (state.activeTask === null || state.activeTask.ID !== state.activeTaskID) {
			state.activeTask = state.tasks.filter((task) => task.ID === state.activeTaskID)[0];
		}
	}

	_activeTaskIDFromURL(uri) {
		if (!uri.fragment) return null;
		return parseInt(uri.fragment.replace(/^T/, ""), 10);
	}

	_activeTask() {
		return this.state.tasks.filter((task) => task.ID === this.state.activeTaskID)[0];
	}

	render() {
		return (
			<div>
				<ul className="nav nav-tabs" role="tablist">
					{this.state.tasks.map((task, i) =>
						<li key={task.ID}
							role="presentation"
							className={task.ID === this.state.activeTaskID ? "active" : ""}>
							<a role="tab"
								href={`#T${task.ID}`}>
								<span className={taskClass(task).text}>
									<i className={taskClass(task).icon}></i> {task.Label}
								</span>
							</a>
						</li>
					)}
				</ul>
				<div className="tab-content">
					<div role="tabpanel" className="tab-pane active">
						{this._activeTask() ?
							<BuildTask task={this._activeTask()} logs={this.state.logs} /> :
							"Loading..."}
					</div>
				</div>
			</div>
		);
	}
}

BuildTasks.propTypes = {
	location: React.PropTypes.string.isRequired,
	tasks: React.PropTypes.array.isRequired,
	logs: React.PropTypes.object.isRequired,
};

function taskClass(task) {
	if (!task.StartedAt) {
		return {icon: "fa fa-circle-o-notch", text: "text-default"};
	}
	if (task.StartedAt && !task.EndedAt) {
		return {icon: "fa fa-circle-o-notch fa-spin", text: "text-info"};
	}
	if (task.Success) {
		return {icon: "fa fa-check", text: "text-success"};
	}
	return {icon: "fa fa-exclamation-circle", text: "text-danger"};
}

export default BuildTasks;
