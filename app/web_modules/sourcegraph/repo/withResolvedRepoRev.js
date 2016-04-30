// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import Dispatcher from "sourcegraph/Dispatcher";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo";
import {urlToRepo} from "sourcegraph/repo/routes";

export default function withResolvedRepoRev(Component) {
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
			state.rev = repoRev(repoSplat);

			state.repoResolution = RepoStore.resolutions.get(state.repo);
			state.repoObj = RepoStore.repos.get(state.repo);
			if (!state.rev) state.rev = state.repoObj && state.repoObj.DefaultBranch || null;

			state.inventory = RepoStore.inventory.get(state.repo, state.rev);
			state.isCloning = RepoStore.repos.isCloning(state.repo);
		}

		onStateTransition(prevState, nextState) {
			if (nextState.repoResolution && prevState.repoResolution !== nextState.repoResolution) {
				if (nextState.repoResolution.Error) {
					// Do nothing.
				} else if (nextState.repoResolution.Result.RemoteRepo) {
					let canonicalPath = `github.com/${nextState.repoResolution.Result.RemoteRepo.Owner}/${nextState.repoResolution.Result.RemoteRepo.Name}`;
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
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
