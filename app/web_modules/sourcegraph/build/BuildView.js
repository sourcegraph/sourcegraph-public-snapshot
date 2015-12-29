import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildBackend from "sourcegraph/build/BuildBackend";
import BuildHeader from "sourcegraph/build/BuildHeader";
import BuildStore from "sourcegraph/build/BuildStore";
import BuildTasks from "sourcegraph/build/BuildTasks";
import BuildTask from "sourcegraph/build/BuildTask";
import LocationAdaptor from "sourcegraph/LocationAdaptor";

const updateIntervalMsec = 1500;

// TODO(sqs): We broke the native CI changeset into
// chunks. Some parts of BuildView are only intended for when
// native CI is enabled. Feature-flag these off until native
// CI is enabled, and then we can remove the non-native-CI
// code paths.
let nativeCIEnabled = (!document || !document.head || !document.head.dataset || document.head.dataset.nativeCiEnabled !== "false");
BuildBackend.nativeCIEnabled = nativeCIEnabled;
let fakeTask;
if (!nativeCIEnabled) {
	fakeTask = {
		ID: 0,
		Label: "Main",
	};
}

class BuildView extends Container {
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
		Dispatcher.asyncDispatch(new BuildActions.WantBuild(this.state.build.Repo, this.state.build.ID, true));
		Dispatcher.asyncDispatch(new BuildActions.WantTasks(this.state.build.Repo, this.state.build.ID, true));
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
			Dispatcher.asyncDispatch(new BuildActions.WantBuild(nextState.build.Repo, nextState.build.ID));
			Dispatcher.asyncDispatch(new BuildActions.WantTasks(nextState.build.Repo, nextState.build.ID));
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

		if (!nativeCIEnabled) {
			fakeTask.Build = {Repo: {URI: "spans"}, ID: this.state.build.ID};
			fakeTask.StartedAt = build.StartedAt;
			fakeTask.EndedAt = build.EndedAt;
			fakeTask.Success = build.Success;
			fakeTask.Failure = build.Failure;
		}

		return (
			<div>
				<BuildHeader build={build} commit={this.state.commit} />
				{!nativeCIEnabled ? <BuildTask task={fakeTask} logs={this.state.logs} /> : null}

				{(nativeCIEnabled && tasks) ? <LocationAdaptor component={BuildTasks} tasks={tasks} logs={this.state.logs} /> : null}
			</div>
		);
	}
}

BuildView.propTypes = {
	build: React.PropTypes.object.isRequired,
	commit: React.PropTypes.object.isRequired,
};

export default BuildView;
