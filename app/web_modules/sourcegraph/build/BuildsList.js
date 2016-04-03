import React from "react";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import {buildStatus, buildClass, elapsed} from "sourcegraph/build/Build";
import {urlToBuild} from "sourcegraph/build/routes";

import {Button} from "sourcegraph/components";

import TimeAgo from "sourcegraph/util/TimeAgo";

import CSSModules from "react-css-modules";
import styles from "./styles/Build.css";

const updateIntervalMsec = 30000;

class BuildsList extends Container {
	static propTypes = {
		params: React.PropTypes.object.isRequired,
	};

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
		Dispatcher.Backends.dispatch(new BuildActions.WantBuilds(this.state.repo, this.state.search, true));
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repo = props.params.splat || null; // null if serving global builds view
		state.search = state.location.search || "?Direction=desc&Sort=updated_at";

		state.builds = BuildStore.buildLists.get(state.repo, state.search);
	}

	stores() { return [BuildStore]; }

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.search !== nextState.search) {
			Dispatcher.Backends.dispatch(new BuildActions.WantBuilds(nextState.repo, nextState.search));
		}
	}

	_rowStatus(build) {
		if (build.Success) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.EndedAt} />];
		if (build.Failure) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.EndedAt} />];
		if (build.Killed) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.EndedAt} />];
		if (build.Queued) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.Started} />];
	}

	render() {
		// TODO(john): show active filter state.
		return (
			<div styleName="list-container">
				<div styleName="list-filters">
					<div styleName="filter-button"><Button outline={true} small={true}><Link to={`${this.state.location.pathname}?Direction=desc&Sort=updated_at`}>All</Link></Button></div>
					<div styleName="filter-button"><Button outline={true} small={true}><Link to={`${this.state.location.pathname}?Direction=desc&Queued=true&Sort=priority`}>Priority Queue</Link></Button></div>
					<div styleName="filter-button"><Button outline={true} small={true}><Link to={`${this.state.location.pathname}?Direction=desc&Active=true&Sort=updated_at`}>Active</Link></Button></div>
					<div styleName="filter-button"><Button outline={true} small={true}><Link to={`${this.state.location.pathname}?Direction=desc&Ended=true&Sort=updated_at`}>Ended</Link></Button></div>
					<div styleName="filter-button"><Button outline={true} small={true}><Link to={`${this.state.location.pathname}?Direction=desc&Sort=updated_at&Succeeded=true`}>Succeeded</Link></Button></div>
					<div styleName="filter-button"><Button outline={true} small={true}><Link to={`${this.state.location.pathname}?Direction=desc&Failed=true&Sort=updated_at`}>Failed</Link></Button></div>
				</div>
				{this.state.builds && [
					<div key="header" styleName="list-item">
						<span styleName="list-id">#</span>
						<span styleName="list-status">Status</span>
						<span styleName="list-elapsed">Elapsed</span>
						<span styleName="list-link"></span>
					</div>,
					...this.state.builds.map((build, i) =>
						<div key={i} styleName={`list-item ${buildClass(build)}`}>
							<span styleName="list-id"><Link to={urlToBuild(build.Repo, build.ID)}>{`${build.ID}`}</Link></span>
							<span styleName="list-status">{this._rowStatus(build)}</span>
							<span styleName="list-elapsed">{elapsed(build)}</span>
							<span styleName="list-link"><Button outline={true} small={true}><Link to={urlToBuild(build.Repo, build.ID)}>View</Link></Button></span>
						</div>
				)]}
				{!this.state.builds && <div styleName="list-empty-state">Sorry, we didn't find any builds.</div>}
			</div>
		);
	}
}

export default CSSModules(BuildsList, styles, {allowMultiple: true});
