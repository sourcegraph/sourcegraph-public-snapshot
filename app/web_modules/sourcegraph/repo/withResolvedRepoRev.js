// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import RepoStore from "sourcegraph/repo/RepoStore";
import TreeStore from "sourcegraph/tree/TreeStore";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import Dispatcher from "sourcegraph/Dispatcher";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo";
import {urlToRepo} from "sourcegraph/repo/routes";
import {rel} from "sourcegraph/app/routePatterns";

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
			return [RepoStore, TreeStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat);

			state.repoResolution = RepoStore.resolutions.get(state.repo);
			state.repoObj = RepoStore.repos.get(state.repo);
			if (!state.rev) {
				state.branch = state.repoObj && state.repoObj.DefaultBranch || null;
				let version = TreeStore.srclibDataVersions.get(state.repo, state.branch);
				// If a revision is not provided, it needs to be resolved. The way we
				// do this depends on the route:
				//
				// - For def/* routes, resolve to the latest built commit, so URLs to a
				// def without a revision won't result in a 404 if the default branch
				// is not built.
				// - For all other routes, resolve to the default branch to show the
				// latest content.
				let paths = this.props.routes ? this.props.routes.map((r) => r.path) : [];
				if (paths.indexOf(rel.def) !== -1 || paths.some((p) => p && p.startsWith(`${rel.def}/-/`))) {
					state.rev = version ? version.CommitID : null;
				} else {
					state.rev = state.repoObj && state.repoObj.DefaultBranch ? state.repoObj.DefaultBranch : null;
				}
			}
			state.inventory = RepoStore.inventory.get(state.repo, state.rev);
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
						this.context.router.replace(urlToRepo(canonicalPath));
						return;
					}

					// If it's a remote repo, do nothing; RepoMain should clone the repository.
				} else if (nextState.repoResolution.Result.Repo) {
					let canonicalPath = nextState.repoResolution.Result.Repo.URI;
					if (nextState.repo !== canonicalPath) {
						this.context.router.replace(urlToRepo(canonicalPath));
						return;
					}

					// Fetch it if it's a local repo.
					Dispatcher.Backends.dispatch(new RepoActions.WantRepo(nextState.repo));
				}
			}
			if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev) {
				if (nextState.repoObj && !nextState.repoObj.Error && !nextState.isCloning && nextState.rev) {
					Dispatcher.Backends.dispatch(new RepoActions.WantInventory(nextState.repo, nextState.rev));
				}
			}
			if (!nextState.rev && nextState.branch) {
				Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.branch));
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
