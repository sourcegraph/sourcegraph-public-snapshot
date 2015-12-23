import React from "react";
import classNames from "classnames";

import {taskClass} from "sourcegraph/build/Build";
import Component from "sourcegraph/Component";

class BuildNav extends Component {
	constructor(props) {
		super(props);
		this.state = {
			topLevelTasks: [],
			subtasks: {},
		};
	}

	reconcileState(state, props) {
		if (state.tasks !== props.tasks) {
			state.tasks = props.tasks;
			state.topLevelTasks = state.tasks.filter((task) => !task.ParentID);

			// Generate subtasks map.
			state.subtasks = {};
			state.tasks.forEach((task) => {
				state.subtasks[task.ID] = state.tasks.filter((subtask) => subtask.ParentID === task.ID);
			});
		}
	}

	render() {
		return (
			<div className="build-nav">
				<ol className="list-unstyled">
				{this.state.topLevelTasks.map((task) => {
					let cls = {
						[`${taskClass(task).text}`]: true,
					};
					return (
						<li key={task.ID}>
							<a className={classNames(cls)} href={`#T${task.ID}`}>
								<i className={taskClass(task).icon}></i>&nbsp;
								<strong>{task.Label}</strong>
							</a>
							<ol className="list-unstyled">
								{this.state.subtasks[task.ID].map((subtask) => (
									<li key={subtask.ID}>
										<a className={classNames({[`${taskClass(subtask).text}`]: true})} href={`#T${subtask.ID}`}>
										<i className={taskClass(subtask).icon}></i>&nbsp;
										{subtask.Label}
									</a>
									</li>
								))}
							</ol>
						</li>
					);
				})}
				</ol>
			</div>
		);
	}
}

BuildNav.propTypes = {
	tasks: React.PropTypes.array.isRequired,
};

export default BuildNav;
