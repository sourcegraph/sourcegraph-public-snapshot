// tslint:disable: typedef ordered-imports

import * as React from "react";
import {DefStore} from "sourcegraph/def/DefStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Helper} from "sourcegraph/blob/BlobLoader";
import {BlobStore, keyForFile} from "sourcegraph/blob/BlobStore";
import {fileLines} from "sourcegraph/util/fileLines";
import {lineFromByte} from "sourcegraph/blob/lineFromByte";
import {computeLineStartBytes} from "sourcegraph/blob/lineFromByte";
import {Header} from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";
import Helmet from "react-helmet";

type Props = any;

type State = any;

// withDefAndRefLocations fetches the def and ref locations for the
// def specified in the properties.
export const withDefAndRefLocations = ({
	reconcileState(state: State, props: Props) {
		state.def = props.params ? props.params.splat[1] : null;
		state.defObj = state.def && state.commitID ? DefStore.defs.get(state.repo, state.commitID, state.def) : null;
		state.defPos = state.def && state.commitID ? DefStore.defs.getPos(state.repo, state.commitID, state.def) : null;
		state.blob = state.path && state.commitID ? (BlobStore.files[keyForFile(state.repo, state.commitID, state.path)] || null) : null;
		state.refLocations = state.def && state.commitID ? DefStore.getRefLocations({
			repo: state.repo,
			commitID: state.commitID,
			def: state.def,
			repos: [],
		}) : null;

		if (!state.refLocations && state.defPos && state.blob) {
			// Compute line start byte where current Def sits.
			let lines = fileLines(state.blob.ContentsString);
			let lineNumber = lineFromByte(lines, state.defPos.DefStart);
			let lineStartBytes = computeLineStartBytes(lines);

			// TODO: remove second argument after we completelt abandon srclib data.
			Dispatcher.Backends.dispatch(new DefActions.WantLocalRefLocations({
				repo: state.repo, commit: state.commitID,
				file: state.defPos.File, line: lineNumber - 1, character: state.defPos.DefStart - lineStartBytes[lineNumber - 1],
			}, {
				repo: state.repo, commitID: state.commitID, def: state.def, repos: [], page: 1,
			}));
		}
	},

	onStateTransition(prevState: State, nextState: State) {
		// Handle change in params OR lost def data (due to auth change, etc.).
		if (nextState.commitID && nextState.def && (nextState.repo !== prevState.repo || nextState.commitID !== prevState.commitID || nextState.def !== prevState.def || (!nextState.defObj && nextState.defObj !== prevState.defObj) || (!nextState.refLocations && nextState.refLocations !== prevState.refLocations))) {
			if (!nextState.defObj) {
				Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.commitID, nextState.def));
			}
		}
	},

	render(state) {
		if (state.defObj && state.defObj.Error) {
			return (
				<div>
					<Helmet title={"Not Found"} />
					<Header
						title={`${httpStatusCode(state.defObj.Error)}`}
						subtitle={`Definition is not available.`} />
				</div>
			);
		}
		return null;
	},

	store: DefStore,
} as Helper);
