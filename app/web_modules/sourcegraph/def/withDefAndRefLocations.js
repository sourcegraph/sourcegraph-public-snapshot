// @flow

import React from "react";
import DefStore from "sourcegraph/def/DefStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";
import type {Helper} from "sourcegraph/blob/BlobLoader";
import Header from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/app/status";

// withDefAndRefLocations fetches the def and ref locations for the
// def specified in the properties.
export default ({
	reconcileState(state, props) {
		state.def = props.params ? props.params.splat[1] : null;

		state.defObj = state.def ? DefStore.defs.get(state.repo, state.rev, state.def) : null;
		state.refLocations = state.def ? DefStore.refLocations.get(state.repo, state.rev, state.def) : null;
	},

	onStateTransition(prevState, nextState, context) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.rev, nextState.def));
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations(nextState.repo, nextState.rev, nextState.def, true));
		}

		if (nextState.defObj && prevState.defObj !== nextState.defObj) {
			context.status.error(nextState.defObj.Error);
		}

		if (nextState.refLocations && prevState.refLocations !== nextState.refLocations) {
			context.status.error(nextState.refLocations.Error);
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
