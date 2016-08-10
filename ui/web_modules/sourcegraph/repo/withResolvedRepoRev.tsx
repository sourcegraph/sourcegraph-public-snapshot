// tslint:disable: typedef ordered-imports curly

import * as React from "react";

import {Container} from "sourcegraph/Container";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo/index";

// withResolvedRepoRev reads the repo, rev, repo resolution, etc.,
// from the route params. If isMainComponent is true, then it also dispatches
// actions to populate that data if necessary (dispatch should only be
// true for the main component, not the nav or other secondary components,
// or else duplicate WantResolveRepo, etc., actions will be dispatched
// and could lead to multiple WantCreateRepo, etc., actions being sent
// to the server).
export function withResolvedRepoRev(Component, isMainComponent?: boolean) {
	type Props = {
		params: any,
		location: HistoryModule.Location,
	};

	isMainComponent = Boolean(isMainComponent);
	class WithResolvedRepoRev extends Container<Props, any> {
		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		_cloningInterval: any;
		_cloningTimeout: any;

		stores() {
			return [RepoStore];
		}

		componentWillUnmount() {
			if (super.componentWillUnmount) super.componentWillUnmount();
			if (this._cloningInterval) clearInterval(this._cloningInterval);
		}

		reconcileState(state, props) {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat); // the original rev from the URL

			state.repoResolution = RepoStore.resolutions.get(state.repo);
			state.repoID = state.repoResolution && !state.repoResolution.Error && state.repoResolution.Repo ? state.repoResolution.Repo : null;
			state.repoObj = RepoStore.repos.get(state.repo);

			state.resolvedRev = state.repoObj && !state.repoObj.Error ? RepoStore.resolvedRevs.get(state.repo, state.rev) : null;
			state.commitID = state.resolvedRev && !state.resolvedRev.Error ? state.resolvedRev.CommitID : null;
			state.inventory = state.commitID ? RepoStore.inventory.get(state.repo, state.commitID) : null;
			state.isCloning = RepoStore.repos.isCloning(state.repo);
		}

		onStateTransition(prevState, nextState) {
			if (!isMainComponent) return;

			// Handle change in params OR lost resolution (due to auth change, etc.).
			if (nextState.repo && !nextState.repoResolution && (prevState.repo !== nextState.repo || prevState.repoResolution !== nextState.repoResolution)) {
				Dispatcher.Backends.dispatch(new RepoActions.WantResolveRepo(nextState.repo));
			}

			if (nextState.repoResolution && prevState.repoResolution !== nextState.repoResolution) {
				if (nextState.repoResolution.Error) {
					// Do nothing.
				} else if (nextState.repoResolution.Repo) {
					let canonicalPath = nextState.repoResolution.CanonicalPath;
					if (nextState.repo !== canonicalPath) {
						let canonicalURL = this.props.location.pathname.replace(new RegExp(this.state.repo, "g"), canonicalPath);
						(this.context as any).router.replace(canonicalURL);
						return;
					}

					// Fetch it if it's a local repo.
					Dispatcher.Backends.dispatch(new RepoActions.WantRepo(nextState.repo));
				} else if (nextState.repoResolution.RemoteRepo) {
					let remoteRepo = nextState.repoResolution.RemoteRepo;
					let canonicalPath = `github.com/${nextState.repoResolution.RemoteRepo.Owner}/${nextState.repoResolution.RemoteRepo.Name}`;
					if (remoteRepo.HTTPCloneURL && !remoteRepo.HTTPCloneURL.startsWith("https://github.com/")) {
						if (remoteRepo.HTTPCloneURL.startsWith("https://")) {
							canonicalPath = remoteRepo.HTTPCloneURL.substr("https://".length);
						} else if (remoteRepo.HTTPCloneURL.startsWith("http://")) {
							canonicalPath = remoteRepo.HTTPCloneURL.substr("http://".length);
						}
					}

					if (nextState.repo !== canonicalPath) {
						let canonicalURL = this.props.location.pathname.replace(new RegExp(this.state.repo, "g"), canonicalPath);
						(this.context as any).router.replace(canonicalURL);
						return;
					}

					// If it's a remote repo, do nothing; RepoMain should clone the repository.
				}
			}
			if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.repoObj !== nextState.repoObj) {
				if (!nextState.commitID && nextState.repoObj && !nextState.repoObj.Error && !nextState.isCloning) {
					Dispatcher.Backends.dispatch(new RepoActions.WantResolveRev(nextState.repo, nextState.rev));
				}
			}
			if (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.repoObj !== nextState.repoObj) {
				if (!nextState.inventory && nextState.commitID && nextState.repoObj && !nextState.repoObj.Error && !nextState.isCloning) {
					Dispatcher.Backends.dispatch(new RepoActions.WantInventory(nextState.repo, nextState.commitID));
				}
			}

			// If the repository is cloning, poll against the server for an
			// update periodically.
			if (isMainComponent && nextState.isCloning && !this._cloningInterval && !this._cloningTimeout) {
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
					if (!global.it) { // skip when testing
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
