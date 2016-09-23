import React from "react";
import {render} from "react-dom";
import {Provider} from "react-redux";

import EventLogger from "../../app/analytics/EventLogger";
import * as Actions from "../../app/actions";

import Background from "../../app/components/Background";
import {SearchIcon} from "../../app/components/Icons";
import BlobAnnotator from "../../app/components/BlobAnnotator";
import createStore from "../../app/store/configureStore";

import {parseURL, isGitHubURL, isSourcegraphURL} from "../../app/utils";

let isSearchAppShown = false; // global state indicating whether the search app is visible
let store = createStore({});

function getFileName(info, {isDelta, path}) {
	if (isDelta) {
		const userSelect = info.querySelector(".user-select-contain");
		if (userSelect) {
			return userSelect.title;
		} else if (info.title) {
			return info.title;
		} else {
			return null;
		}
	} else {
		return path;
	}
}

function injectBackgroundApp() {
	if (!document.getElementById("sourcegraph-app-background")) {
		let backgroundContainer = document.createElement("div");
		backgroundContainer.id = "sourcegraph-app-background";
		backgroundContainer.style.display = "none";
		document.body.appendChild(backgroundContainer);
		injectComponent(<Background />, backgroundContainer);
	}
}

function injectBlobAnnotator() {
	if (!isGitHubURL()) return;

	const {user, repo, rev, path, isDelta} = parseURL();
	const files = document.querySelectorAll(".file");
	for (let i = 0; i < files.length; ++i) {
		const file = files[i];
		const info = file.querySelector(".file-info");
		const blob = file.querySelector(".blob-wrapper");
		if (!blob) continue;

		const infoFilePath = getFileName(info, {isDelta, path});
		if (!infoFilePath) continue;

		if (file.dataset && file.dataset["sgAnnotator"]) continue; // prevent injecting twice
		file.dataset["sgAnnotator"] = true;

		const blobAnnotatorContainer = document.createElement("span");
		info.appendChild(blobAnnotatorContainer);
		injectComponent(<BlobAnnotator path={infoFilePath} blobElement={blob} />, blobAnnotatorContainer);
	}
}

function injectComponent(component, mountElement) {
	render(<Provider store={store}>{component}</Provider>, mountElement);
}

function injectModules() {
	injectBackgroundApp();
	injectBlobAnnotator();

	// Add invisible div to the page to indicate injection has completed.
	if (!document.getElementById("sourcegraph-app-bootstrap")) {
		let el = document.createElement("div");
		el.id = "sourcegraph-app-bootstrap";
		el.style.display = "none";
		document.body.appendChild(el);
	}
}

window.addEventListener("load", () => {
	chrome.runtime.sendMessage(null, {type: "get"}, {}, (state) => {
		const accessToken = state.accessToken;
		if (isGitHubURL()) {
			store.dispatch(Actions.setAccessToken(accessToken));
		} else if (isSourcegraphURL()) {
			const regexp = /accessToken\\":\\"([-A-Za-z0-9_.]+)\\"/;
			const matchResult = document.head.innerHTML.match(regexp);
			if (matchResult) store.dispatch(Actions.setAccessToken(matchResult[1]));
		}
		injectModules();
	});
	chrome.runtime.sendMessage(null, {type: "getIdentity"}, {}, (identity) => {
		if (identity) EventLogger.updatePropsForUser(identity);
	});
});
document.addEventListener("pjax:success", () => {
	injectModules();
});

document.addEventListener("sourcegraph:identify", (ev) => {
	if (ev && ev.detail) {
		EventLogger.updatePropsForUser(ev.detail);
		chrome.runtime.sendMessage(null, {type: "setIdentity", identity: ev.detail}, {});
	}
});
