// @flow

import React from "react";
import DefStore from "sourcegraph/def/DefStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";
import type {Helper} from "sourcegraph/blob/BlobLoader";
import Header from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/app/status";

// withDefAndRefs fetches the def and refs for the def specified in
// the properties.
export default ({
	reconcileState(state, props) {
		state.def = props.params ? props.params.splat[1] : null;

		state.defObj = state.def ? DefStore.defs.get(state.repo, state.rev, state.def) : null;
		state.refs = state.def ? DefStore.refs.get(state.repo, state.rev, state.def) : null;
	},

	onStateTransition(prevState, nextState, context) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.rev, nextState.def));
			Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def));
		}

		if (nextState.defObj && prevState.defObj !== nextState.defObj) {
			context.status.error(nextState.defObj.Error);
		}

		if (nextState.refs && prevState.refs !== nextState.refs) {
			context.status.error(nextState.refs.Error);
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
