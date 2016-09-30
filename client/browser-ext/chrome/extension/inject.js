import React from "react";
import {render} from "react-dom";
import {unmountComponentAtNode} from "react-dom";
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
		} else if (info.tagName === "A") {
			return info.innerHTML.trim(); // get text and strip whitespace
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

		const blobAnnotatorContainer = document.createElement("span");
		blobAnnotatorContainer.className = "sourcegraph-app-annotator";
		info.appendChild(blobAnnotatorContainer);
		injectComponent(<BlobAnnotator path={infoFilePath} blobElement={blob} infoElement={info} />, blobAnnotatorContainer);
	}
}

function ejectComponent(mountElement) {
	unmountComponentAtNode(mountElement);
	mountElement.remove();
}

function injectComponent(component, mountElement) {
	render(<Provider store={store}>{component}</Provider>, mountElement);
}

function ejectModules() {
	var annotators = document.getElementsByClassName('sourcegraph-app-annotator');
	var background = document.getElementById('sourcegraph-app-background');
	var bootstrap = document.getElementById('sourcegraph-app-bootstrap');

	for (let idx = annotators.length - 1; idx >= 0; idx--) {
		ejectComponent(annotators.item(idx));
	}

	if (background)
		ejectComponent(background);

	if (bootstrap)
		bootstrap.remove(); // Not a react component
}

function injectModules() {
	// Add invisible div to the page to indicate injection has completed.
	if (!document.getElementById("sourcegraph-app-bootstrap")) {
		injectBackgroundApp();
		injectBlobAnnotator();

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

document.addEventListener("pjax:end", () => {
	// Unmount and remount react components because pjax breaks
	// dynamically registered event handlers like mouseover/click etc..
	ejectModules();
	injectModules();
});

document.addEventListener("sourcegraph:identify", (ev) => {
	if (ev && ev.detail) {
		EventLogger.updatePropsForUser(ev.detail);
		chrome.runtime.sendMessage(null, {type: "setIdentity", identity: ev.detail}, {});
	}
});
