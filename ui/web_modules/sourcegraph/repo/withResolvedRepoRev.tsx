import * as React from "react";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Location} from "sourcegraph/Location";
import {repoParam, repoPath, repoRev} from "sourcegraph/repo";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {Store} from "sourcegraph/Store";

interface Props {
	params: any;
	location: Location;
}

type State = any;

// withResolvedRepoRev reads the repo, rev, etc.,
export function withResolvedRepoRev(Component: any): React.ComponentClass<Props> {

	class WithResolvedRepoRev extends Container<Props, State> {
		stores(): Store<any>[] {
			return [RepoStore];
		}

		componentWillUnmount(): void {
			if (super.componentWillUnmount) {
				super.componentWillUnmount();
			}
		}

		reconcileState(state: State, props: Props): void {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat); // the original rev from the URL

			state.repoObj = RepoStore.repos.get(state.repo);
		}

		onStateTransition(prevState: State, nextState: State): void {
			if (nextState.repo && !nextState.repoObj && (prevState.repo !== nextState.repo)) {
				Dispatcher.Backends.dispatch(new RepoActions.WantRepo(nextState.repo));
			}
		}

		render(): JSX.Element | null {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
