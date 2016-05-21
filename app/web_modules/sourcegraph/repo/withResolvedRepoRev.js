// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import Dispatcher from "sourcegraph/Dispatcher";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo";

// withResolvedRepoRev reads the repo, rev, repo resolution, etc.,
// from the route params. If isMainComponent is true, then it also dispatches
// actions to populate that data if necessary (dispatch should only be
// true for the main component, not the nav or other secondary components,
// or else duplicate WantResolveRepo, etc., actions will be dispatched
// and could lead to multiple WantCreateRepo, etc., actions being sent
// to the server).
export default function withResolvedRepoRev(Component: ReactClass, isMainComponent?: bool): ReactClass {
	isMainComponent = Boolean(isMainComponent);
	class WithResolvedRepoRev extends Container {
		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		static propTypes = {
			params: React.PropTypes.object.isRequired,
		};

		stores() {
			return [RepoStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat); // the original rev from the URL

			state.repoResolution = RepoStore.resolutions.get(state.repo);
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
				} else if (nextState.repoResolution.Result.RemoteRepo) {
					let remoteRepo = nextState.repoResolution.Result.RemoteRepo;
					let canonicalPath = `github.com/${nextState.repoResolution.Result.RemoteRepo.Owner}/${nextState.repoResolution.Result.RemoteRepo.Name}`;
					if (remoteRepo.HTTPCloneURL && !remoteRepo.HTTPCloneURL.startsWith("https://github.com/")) {
						if (remoteRepo.HTTPCloneURL.startsWith("https://")) {
							canonicalPath = remoteRepo.HTTPCloneURL.substr("https://".length);
						} else if (remoteRepo.HTTPCloneURL.startsWith("http://")) {
							canonicalPath = remoteRepo.HTTPCloneURL.substr("http://".length);
						}
					}

					if (nextState.repo !== canonicalPath) {
						let canonicalURL = this.props.location.pathname.replace(new RegExp(this.state.repo, "g"), canonicalPath);
						this.context.router.replace(canonicalURL);
						return;
					}

					// If it's a remote repo, do nothing; RepoMain should clone the repository.
				} else if (nextState.repoResolution.Result.Repo) {
					let canonicalPath = nextState.repoResolution.Result.Repo.URI;
					if (nextState.repo !== canonicalPath) {
						let canonicalURL = this.props.location.pathname.replace(new RegExp(this.state.repo, "g"), canonicalPath);
						this.context.router.replace(canonicalURL);
						return;
					}

					// Fetch it if it's a local repo.
					Dispatcher.Backends.dispatch(new RepoActions.WantRepo(nextState.repo));
				}
			}
			if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.repoObj !== nextState.repoObj) {
				if (nextState.repoObj && !nextState.repoObj.Error && !nextState.isCloning) {
					Dispatcher.Backends.dispatch(new RepoActions.WantResolveRev(nextState.repo, nextState.rev));
				}
			}
			if (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.repoObj !== nextState.repoObj) {
				if (nextState.commitID && nextState.repoObj && !nextState.repoObj.Error && !nextState.isCloning) {
					Dispatcher.Backends.dispatch(new RepoActions.WantInventory(nextState.repo, nextState.commitID));
				}
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
