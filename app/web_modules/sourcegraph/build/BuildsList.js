import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import {buildStatus, buildClass, elapsed} from "sourcegraph/build/Build";
import {trimRepo} from "sourcegraph/repo";
import {urlToBuild} from "sourcegraph/build/routes";
import {urlToRepo} from "sourcegraph/repo/routes";

import {Button} from "sourcegraph/components";

import TimeAgo from "sourcegraph/util/TimeAgo";

import CSSModules from "react-css-modules";
import styles from "./styles/Build.css";
import btnStyles from "sourcegraph/components/styles/button.css";

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
		Dispatcher.Backends.dispatch(new BuildActions.WantBuilds(this.state.repo, this._translateQuery(this.state.search), true));
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repo = props.params.splat || null; // null if serving global builds view
		state.search = state.location.search;

		state.builds = BuildStore.buildLists.get(state.repo, this._translateQuery(state.search));
	}

	stores() { return [BuildStore]; }

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.search !== nextState.search) {
			Dispatcher.Backends.dispatch(new BuildActions.WantBuilds(nextState.repo, this._translateQuery(nextState.search)));
		}
	}

	_translateQuery(search) {
		switch (search) {
		case "?filter=all": return "?Direction=desc&Sort=updated_at";
		case "?filter=priority": return "?Direction=desc&Queued=true&Sort=priority";
		case "?filter=active": return "?Direction=desc&Active=true&Sort=updated_at";
		case "?filter=ended": return "?Direction=desc&Ended=true&Sort=updated_at";
		case "?filter=succeeded": return "?Direction=desc&Sort=updated_at&Succeeded=true";
		case "?filter=failed": return "?Direction=desc&Failed=true&Sort=updated_at";
		default: return "?Direction=desc&Sort=updated_at";
		}
	}

	_rowStatus(build) {
		if (build.Success) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.EndedAt} />];
		if (build.Failure) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.EndedAt} />];
		if (build.Killed) return [`${buildStatus(build)} `, <TimeAgo key="time" time={build.EndedAt} />];
		if (build.Queue) return [`${buildStatus(build)} `, (build.StartedAt && <TimeAgo key="time" time={build.StartedAt} />)];
		return null;
	}

	render() {
		return (
			<div styleName="list-container">
				<Helmet title={`Builds | ${this.state.repo ? trimRepo(this.state.repo) : ""}`} />
				<div styleName="list-header">Builds</div>
				<div styleName="list-filters">
					<div styleName="filter-button">
						<Link activeClassName={btnStyles["outline-active"]} to={`${this.state.location.pathname}?filter=all`}><Button size="small" outline={true}>All</Button></Link>
					</div>
					<div styleName="filter-button">
						<Link activeClassName={btnStyles["outline-active"]} to={`${this.state.location.pathname}?filter=priority`}><Button size="small" outline={true}>Priority Queue</Button></Link>
					</div>
					<div styleName="filter-button">
						<Link activeClassName={btnStyles["outline-active"]} to={`${this.state.location.pathname}?filter=active`}><Button size="small" outline={true}>Active</Button></Link>
					</div>
					<div styleName="filter-button">
						<Link activeClassName={btnStyles["outline-active"]} to={`${this.state.location.pathname}?filter=ended`}><Button size="small" outline={true}>Ended</Button></Link>
					</div>
					<div styleName="filter-button">
						<Link activeClassName={btnStyles["outline-active"]} to={`${this.state.location.pathname}?filter=succeeded`}><Button size="small" outline={true}>Succeeded</Button></Link>
					</div>
					<div styleName="filter-button">
						<Link activeClassName={btnStyles["outline-active"]} to={`${this.state.location.pathname}?filter=failed`}><Button size="small" outline={true}>Failed</Button></Link>
					</div>
				</div>
				{this.state.builds !== null && this.state.builds.length !== 0 && [
					<div key="header" styleName="list-item">
						<span styleName="list-id">#</span>
						<span styleName="list-repo">Repository</span>
						<span styleName="list-status">Status</span>
						<span styleName="list-elapsed">Elapsed</span>
					</div>,
					...this.state.builds.map((build, i) =>
						<div key={i} styleName={`list-item ${buildClass(build)}`}>
							<span styleName="list-id">
								{!build.StartedAt &&
									<span>{build.ID}</span>
								}
								{build.StartedAt &&
									<Link to={urlToBuild(this.state.repo || build.RepoPath, build.ID)}><Button size="small" block={true} outline={true}>{`${build.ID}`}</Button></Link>
								}
							</span>
							<span styleName="list-repo"><a href={urlToRepo(this.state.repo || build.RepoPath)}>{this.state.repo || build.RepoPath}</a></span>
							<span styleName="list-status">{this._rowStatus(build)}</span>
							<span styleName="list-elapsed">{elapsed(build)}</span>
						</div>
				)]}
				{this.state.builds !== null && this.state.builds.length === 0 && <div styleName="list-empty-state">Sorry, we didn't find any builds.</div>}
			</div>
		);
	}
}

export default CSSModules(BuildsList, styles, {allowMultiple: true});
