// @flow weak

import React from "react";
import DefPopup from "sourcegraph/def/DefPopup";
import type {Helper} from "sourcegraph/blob/BlobLoader";

// blobWithDefBox uses the def's path as the blob file to load, and it
// passes a DefPopup child to be displayed in the blob margin.
export default ({
	reconcileState(state, props) {
		state.path = state.defObj ? state.defObj.File : null;
	},

	renderProps(state) {
		return state.defObj ? {children: <DefPopup def={state.defObj} refs={state.refs} path={state.defObj.File} byte={state.defObj.DefStart} />} : null;
	},
}: Helper);
