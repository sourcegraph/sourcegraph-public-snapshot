// @flow

import React from "react";
import DefStore from "sourcegraph/def/DefStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";
import type {Helper} from "sourcegraph/blob/BlobLoader";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";

// withDefAndRefLocations fetches the def and ref locations for the
// def specified in the properties.
export default ({
	reconcileState(state, props) {
		state.def = props.params ? props.params.splat[1] : null;
		state.defObj = state.def ? DefStore.defs.get(state.repo, state.rev, state.def) : null;
		state.refLocations = state.def ? DefStore.getRefLocations({
			repo: state.repo,
			rev: state.rev,
			def: state.def,
			repos: [],
		}) : null;
	},

	onStateTransition(prevState, nextState) {
		// Handle change in params OR lost def data (due to auth change, etc.).
		if (nextState.def && (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def || nextState.defObj !== prevState.defObj || nextState.refLocations !== prevState.refLocations)) {
			if (!nextState.defObj) {
				Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.rev, nextState.def));
			}
			if (!nextState.refLocations) {
				Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
					repo: nextState.repo, rev: nextState.rev, def: nextState.def, repos: [],
				}));
			}
		}
	},

	render(state) {
		if (state.defObj && state.defObj.Error) {
			return (
				<Header
					title={`${httpStatusCode(state.defObj.Error)}`}
					subtitle={`Definition is not available.`} />
			);
		}
		return null;
	},

	store: DefStore,
}: Helper);
