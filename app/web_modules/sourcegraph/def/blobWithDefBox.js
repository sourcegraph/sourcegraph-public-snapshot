// @flow weak

import React from "react";
import DefPopup from "sourcegraph/def/DefPopup";
import type {Helper} from "sourcegraph/blob/BlobLoader";
import LinkGitHubCTA from "sourcegraph/def/LinkGitHubCTA";

// blobWithDefBox uses the def's path as the blob file to load, and it
// passes a DefPopup child to be displayed in the blob margin.
export default ({
	reconcileState(state, props) {
		state.path = state.defObj && !state.defObj.Error ? state.defObj.File : null;
	},

	renderProps(state) {
		return state.defObj && !state.defObj.Error ? {children: <DefPopup def={state.defObj} refLocations={state.refLocations} path={state.defObj.File} byte={state.defObj.DefStart} onboardingCTA={<LinkGitHubCTA/>} />} : null;
	},
}: Helper);
