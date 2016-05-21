// @flow weak

import React from "react";
import DefStore from "sourcegraph/def/DefStore";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";
import Container from "sourcegraph/Container";
import {routeParams as defRouteParams} from "sourcegraph/def";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";

// withDef fetches the def specified in the params. It also fetches
// the def stored in DefStore.highlightedDef.
export default function withDef(Component) {
	class WithDef extends Container {
		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			rev: React.PropTypes.string,
			params: React.PropTypes.object.isRequired,
			isCloning: React.PropTypes.bool,
		};

		stores() { return [DefStore, RepoStore]; }

		reconcileState(state, props) {
			Object.assign(state, props);

			if (!props.def) state.def = props.params ? props.params.splat[1] : null;
			state.defObj = state.def ? DefStore.defs.get(state.repo, state.rev, state.def) : null;
			state.commitID = state.defObj && !state.defObj.Error ? state.defObj.CommitID : null;

			state.highlightedDef = DefStore.highlightedDef || null;
			if (state.highlightedDef) {
				let {repo, rev, def} = defRouteParams(state.highlightedDef);
				state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
			} else {
				state.highlightedDefObj = null;
			}

			state.isCloning = RepoStore.repos.isCloning(state.repo);
		}

		onStateTransition(prevState, nextState) {
			if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
				Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.rev, nextState.def));
			}

			if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
				let {repo, rev, def} = defRouteParams(nextState.highlightedDef);
				Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
			}
		}

		render() {
			if (this.state.isCloning) {
				return (
					<Header
						title="Cloning this repository"
						subtitle="Refresh this page in a minute." />
				);
			}

			if (this.state.defObj && this.state.defObj.Error) {
				return (
					<Header
						title={`${httpStatusCode(this.state.defObj.Error)}`}
						subtitle={`Definition is not available.`} />
				);
			}

			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithDef;
}
