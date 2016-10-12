import * as React from "react";
import Helmet from "react-helmet";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend"; // for side effects
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {Store} from "sourcegraph/Store";
import {Repos} from "sourcegraph/user/settings/Repos";

const reposQuerystring = "RemoteOnly=true";

interface Props {
	location: any;
}

type State = any;

export class UserSettingsReposMain extends Container<Props, State> {

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.repos = RepoStore.repos.list(reposQuerystring);
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (nextState.repos !== prevState.repos) {
			Dispatcher.Backends.dispatch(new RepoActions.WantRepos(reposQuerystring));
		}
	}

	stores(): Store<any>[] {
		return [RepoStore];
	}

	render(): JSX.Element | null {
		let repos = this.state.repos ? this.state.repos.Repos || [] : null;
		return (
			<div>
				<Helmet title="Repositories" />
				<Repos repos={repos} location={this.props.location} />
			</div>
		);
	}
}
