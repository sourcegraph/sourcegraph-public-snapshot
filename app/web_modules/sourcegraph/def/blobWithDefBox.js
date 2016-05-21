// @flow weak

import React from "react";
import DefPopup from "sourcegraph/def/DefPopup";
import type {Helper} from "sourcegraph/blob/BlobLoader";
import DefStore from "sourcegraph/def/DefStore";

// blobWithDefBox uses the def's path as the blob file to load, and it
// passes a DefPopup child to be displayed in the blob margin.
export default ({
	reconcileState(state, props) {
		const defPos = DefStore.defs.getPos(state.repo, state.rev, state.def);
		state.path = defPos && !defPos.Error ? defPos.File : null;
		state.startByte = defPos && !defPos.Error ? defPos.DefStart : null;
		state.endByte = defPos && !defPos.Error ? defPos.DefEnd : null;

		const def = DefStore.defs.get(state.repo, state.rev, state.def);
		state.commitID = def && !def.Error ? def.CommitID : null;
	},

	renderProps(state) {
		return state.defObj && !state.defObj.Error ? {
			children: <DefPopup
				rev={state.rev}
				def={state.defObj}
				refLocations={state.refLocations}
				path={state.path}
				byte={state.startByte}
				location={state.location} />,
		} : null;
	},
}: Helper);
