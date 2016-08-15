// tslint:disable: typedef ordered-imports

import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import {Commit} from "sourcegraph/vcs/Commit";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import {BuildHeader} from "sourcegraph/build/BuildHeader";
import {BuildStore} from "sourcegraph/build/BuildStore";
import {BuildTasks} from "sourcegraph/build/BuildTasks";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import {urlToBuilds} from "sourcegraph/build/routes";
import {trimRepo} from "sourcegraph/repo/index";

import {Button} from "sourcegraph/components/index";

import * as styles from "sourcegraph/build/styles/Build.css";

const updateIntervalMsec = 1500;

interface Props {
	params: any;
}

type State = any;

export class BuildContainer extends Container<Props, State> {
	static contextTypes = {
		user: React.PropTypes.object,
	};

	_updateIntervalID: any;

	constructor(props: Props) {
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
		Dispatcher.Backends.dispatch(new BuildActions.WantBuild(this.state.repo, this.state.id, true));
		Dispatcher.Backends.dispatch(new BuildActions.WantTasks(this.state.repo, this.state.id, true));
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.repo = props.params.splat;
		state.id = props.params.id;

		state.build = BuildStore.builds.get(state.repo, state.id);
		state.tasks = BuildStore.tasks.get(state.repo, state.id);
		state.commit = state.build ? TreeStore.commits.get(state.repo, state.build.CommitID, "") : null;
		state.logs = BuildStore.logs;
	}

	stores() { return [BuildStore, TreeStore]; }

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState.repo !== nextState.repo || prevState.id !== nextState.id) {
			Dispatcher.Backends.dispatch(new BuildActions.WantBuild(nextState.repo, nextState.id));
			Dispatcher.Backends.dispatch(new BuildActions.WantTasks(nextState.repo, nextState.id));
		}
		if (nextState.build && prevState.build !== nextState.build) {
			Dispatcher.Backends.dispatch(new TreeActions.WantCommit(nextState.repo, nextState.build.CommitID, ""));
		}
	}

	render(): JSX.Element | null {
		if (!this.state.build) {
			return null;
		}

		return (
			<div className={styles.build_container}>
				<Helmet title={`Build #${this.state.id} | ${trimRepo(this.state.repo)}`} />
				<div className={styles.actions}>
					<Link to={urlToBuilds(this.state.repo)}><Button size="large" outline={true}>View All Builds</Button></Link>
					{(this.context as any).user && (this.context as any).user.Admin && <Button style={{marginLeft: "1.5rem"}} size="small" outline={true} onClick={() => {
						Dispatcher.Backends.dispatch(new BuildActions.CreateBuild(this.state.repo, this.state.build.CommitID, this.state.build.Branch, this.state.build.Tag, true));
					}}>Rebuild</Button>}
				</div>
				<BuildHeader build={this.state.build} />
				{this.state.commit && <div className={styles.commit}><Commit commit={this.state.commit} full={false} /></div>}
				{this.state.tasks && this.state.tasks.BuildTasks && <BuildTasks tasks={this.state.tasks.BuildTasks} logs={this.state.logs} />}
			</div>
		);
	}
}
