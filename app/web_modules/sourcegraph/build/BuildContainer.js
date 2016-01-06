import React from "react";

import Commit from "sourcegraph/vcs/Commit";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import "./BuildBackend"; // for side effects
import BuildHeader from "sourcegraph/build/BuildHeader";
import BuildNav from "sourcegraph/build/BuildNav";
import BuildStore from "sourcegraph/build/BuildStore";
import BuildTasks from "sourcegraph/build/BuildTasks";

const updateIntervalMsec = 1500;

class BuildContainer extends Container {
	constructor(props) {
		super(props);
		this._updateIntervalID = null;
	}

	componentDidMount() {
		this._startUpdate();
		super.componentDidMount();
	}

	componentWillUnmount() {
		this._stopUpdate();
		super.componentWillUnmount();
	}

	_startUpdate() {
		if (this._updateIntervalID === null) {
			this._updateIntervalID = setInterval(this._update.bind(this), updateIntervalMsec);
		}
	}

	_stopUpdate() {
		if (this._updateIntervalID !== null) {
			clearInterval(this._updateIntervalID);
			this._updateIntervalID = null;
		}
	}

	_update() {
		Dispatcher.dispatch(new BuildActions.WantBuild(this.state.build.Repo, this.state.build.ID, true));
		Dispatcher.dispatch(new BuildActions.WantTasks(this.state.build.Repo, this.state.build.ID, true));
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.builds = BuildStore.builds;
		state.logs = BuildStore.logs;
		state.tasks = BuildStore.tasks;
	}

	stores() { return [BuildStore]; }

	onStateTransition(prevState, nextState) {
		if (prevState.build !== nextState.build) {
			Dispatcher.dispatch(new BuildActions.WantBuild(nextState.build.Repo, nextState.build.ID));
			Dispatcher.dispatch(new BuildActions.WantTasks(nextState.build.Repo, nextState.build.ID));
		}
	}

	render() {
		let tasks = this.state.tasks.get(this.state.build.Repo, this.state.build.ID);
		if (tasks !== null) {
			tasks = tasks.BuildTasks;
		}

		let build = this.state.builds.get(this.state.build.Repo, this.state.build.ID);
		if (build === null) {
			build = this.state.build;
		}

		return (
			<div className="row build-container">
					<div className="col-md-3 col-lg-2">
						<BuildHeader build={build} commit={this.state.commit} />
						{tasks ? <BuildNav build={build} tasks={tasks} /> : null}
					</div>
					<div className="col-md-9 col-lg-10">
						<Commit commit={this.state.commit} />
						{tasks ? <BuildTasks tasks={tasks} logs={this.state.logs} /> : null}
					</div>
			</div>
		);
	}
}

BuildContainer.propTypes = {
	build: React.PropTypes.object.isRequired,
	commit: React.PropTypes.object.isRequired,
};

export default BuildContainer;
