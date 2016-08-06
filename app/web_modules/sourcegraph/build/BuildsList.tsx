// tslint:disable

import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import {buildStatus, buildClass, elapsed} from "sourcegraph/build/Build";
import {trimRepo} from "sourcegraph/repo/index";
import {urlToBuild} from "sourcegraph/build/routes";
import {urlToRepo} from "sourcegraph/repo/routes";

import {Button} from "sourcegraph/components/index";

import TimeAgo from "sourcegraph/util/TimeAgo";

import CSSModules from "react-css-modules";
import * as styles from "./styles/Build.css";
import * as btnStyles from "sourcegraph/components/styles/button.css";

const updateIntervalMsec = 30000;

class BuildsList extends Container<any, any> {
	_updateIntervalID: any;
	
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

	render(): JSX.Element | null {
		return (
			<div className={styles.list_container}>
				<Helmet title={`Builds | ${this.state.repo ? trimRepo(this.state.repo) : ""}`} />
				<div className={styles.list_header}>Builds</div>
				<div className={styles.list_filters}>
					<div className={styles.filter_button}>
						<Link activeClassName={btnStyles.outline_active} to={`${this.state.location.pathname}?filter=all`}><Button size="small" outline={true}>All</Button></Link>
					</div>
					<div className={styles.filter_button}>
						<Link activeClassName={btnStyles.outline_active} to={`${this.state.location.pathname}?filter=priority`}><Button size="small" outline={true}>Priority Queue</Button></Link>
					</div>
					<div className={styles.filter_button}>
						<Link activeClassName={btnStyles.outline_active} to={`${this.state.location.pathname}?filter=active`}><Button size="small" outline={true}>Active</Button></Link>
					</div>
					<div className={styles.filter_button}>
						<Link activeClassName={btnStyles.outline_active} to={`${this.state.location.pathname}?filter=ended`}><Button size="small" outline={true}>Ended</Button></Link>
					</div>
					<div className={styles.filter_button}>
						<Link activeClassName={btnStyles.outline_active} to={`${this.state.location.pathname}?filter=succeeded`}><Button size="small" outline={true}>Succeeded</Button></Link>
					</div>
					<div className={styles.filter_button}>
						<Link activeClassName={btnStyles.outline_active} to={`${this.state.location.pathname}?filter=failed`}><Button size="small" outline={true}>Failed</Button></Link>
					</div>
				</div>
				{this.state.builds !== null && this.state.builds.length !== 0 && [
					<div key="header" className={styles.list_item}>
						<span className={styles.list_id}>#</span>
						<span className={styles.list_repo}>Repository</span>
						<span className={styles.list_status}>Status</span>
						<span className={styles.list_elapsed}>Elapsed</span>
					</div>,
					...this.state.builds.map((build, i) =>
						<div key={i} styleName={`list_item ${buildClass(build)}`}>
							<span className={styles.list_id}>
								{!build.StartedAt &&
									<span>{build.ID}</span>
								}
								{build.StartedAt &&
									<Link to={urlToBuild(this.state.repo || build.RepoPath, build.ID)}><Button size="small" block={true} outline={true}>{`${build.ID}`}</Button></Link>
								}
							</span>
							<span className={styles.list_repo}><a href={urlToRepo(this.state.repo || build.RepoPath)}>{this.state.repo || build.RepoPath}</a></span>
							<span className={styles.list_status}>{this._rowStatus(build)}</span>
							<span className={styles.list_elapsed}>{elapsed(build)}</span>
						</div>
				)]}
				{this.state.builds !== null && this.state.builds.length === 0 && <div className={styles.list_empty_state}>Sorry, we didn't find any builds.</div>}
			</div>
		);
	}
}

export default CSSModules(BuildsList, styles, {allowMultiple: true});
