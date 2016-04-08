// @flow weak

import React from "react";
import DefStore from "sourcegraph/def/DefStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";
import Container from "sourcegraph/Container";
import {routeParams as defRouteParams} from "sourcegraph/def";

// withDef fetches the def specified in the params. It also fetches
// the def stored in DefStore.highlightedDef.
export default function withDef(Component) {
	class WithDef extends Container {
		static contextTypes = {
			httpResponse: React.PropTypes.object,
		};

		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			rev: React.PropTypes.string.isRequired,
			params: React.PropTypes.object.isRequired,
		};

		stores() { return [DefStore]; }

		reconcileState(state, props) {
			Object.assign(state, props);

			if (!props.def) state.def = props.params ? props.params.splat[1] : null;
			state.defObj = state.def ? DefStore.defs.get(state.repo, state.rev, state.def) : null;

			state.highlightedDef = DefStore.highlightedDef || null;
			if (state.highlightedDef) {
				let {repo, rev, def} = defRouteParams(state.highlightedDef);
				state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
			} else {
				state.highlightedDefObj = null;
			}
		}

		onStateTransition(prevState, nextState) {
			if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
				Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.rev, nextState.def));
			}

			if (nextState.defObj && prevState.defObj !== nextState.defObj) {
				this.context.httpResponse.setStatusCode(nextState.defObj.Error ? 404 : 200);
			}

			if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
				let {repo, rev, def} = defRouteParams(nextState.highlightedDef);
				Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithDef;
}


