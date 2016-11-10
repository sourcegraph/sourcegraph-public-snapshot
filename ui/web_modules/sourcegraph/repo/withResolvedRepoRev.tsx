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
		_cloningInterval: any;
		_cloningTimeout: any;

		stores(): Store<any>[] {
			return [RepoStore];
		}

		componentWillUnmount(): void {
			if (super.componentWillUnmount) {
				super.componentWillUnmount();
			}
			if (this._cloningInterval) {
				clearInterval(this._cloningInterval);
			}
		}

		reconcileState(state: State, props: Props): void {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat); // the original rev from the URL

			state.repoObj = RepoStore.repos.get(state.repo);

			state.resolvedRev = state.repoObj && !state.repoObj.Error ? RepoStore.resolvedRevs.get(state.repo, state.rev) : null;
			state.commitID = state.resolvedRev && !state.resolvedRev.Error ? state.resolvedRev.CommitID : null;
			state.isCloning = RepoStore.repos.isCloning(state.repo);
		}

		onStateTransition(prevState: State, nextState: State): void {
			if (nextState.repo && !nextState.repoObj && (prevState.repo !== nextState.repo)) {
				Dispatcher.Backends.dispatch(new RepoActions.WantRepo(nextState.repo));
			}
			if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.repoObj !== nextState.repoObj) {
				if (!nextState.commitID && nextState.repoObj && !nextState.repoObj.Error && !nextState.isCloning) {
					Dispatcher.Backends.dispatch(new RepoActions.WantResolveRev(nextState.repo, nextState.rev));
				}
			}

			// If the repository is cloning, poll against the server for an
			// update periodically.
			if (nextState.isCloning && !this._cloningInterval && !this._cloningTimeout) {
				// If the cloning would be quick, we don't want to flicker the
				// loading screen, so display the screen for at least 1s.
				const displayForAtLeast = 500;
				const pollInterval = 500;
				const maxAttempts = 10000 / pollInterval; // 10s / 20 times

				this._cloningTimeout = true;
				let attempt = 0;
				setTimeout(() => {
					attempt++;
					if (attempt > maxAttempts) {
						clearInterval(this._cloningInterval);
						this._cloningInterval = null;
					}
					this._cloningTimeout = false;
					if (typeof it === "undefined") { // skip when testing
						this._cloningInterval = setInterval(() => {
							Dispatcher.Backends.dispatch(new RepoActions.WantResolveRev(nextState.repo, nextState.rev, true));
						}, pollInterval);
					}
				}, displayForAtLeast);
			} else if (!nextState.isCloning && this._cloningInterval) {
				clearInterval(this._cloningInterval);
				this._cloningInterval = null;
			}
		}

		render(): JSX.Element | null {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
