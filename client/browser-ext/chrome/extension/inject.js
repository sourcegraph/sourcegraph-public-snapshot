import React from "react";
import {render} from "react-dom";
import {unmountComponentAtNode} from "react-dom";
import {Provider} from "react-redux";

import EventLogger from "../../app/analytics/EventLogger";
import * as Actions from "../../app/actions";

import Background from "../../app/components/Background";
import {SourcegraphIcon} from "../../app/components/Icons";
import BlobAnnotator from "../../app/components/BlobAnnotator";
import createStore from "../../app/store/configureStore";

import {parseURL, isGitHubURL, isSourcegraphURL, readableGitHubRoute, getGitHubRoute, getLinesOfCode} from "../../app/utils";

let store = createStore({});
const isTooBigThreshold = 5000;

function getFileName(info, {isDelta, path}) {
	if (isDelta) {
		const userSelect = info.querySelector(".user-select-contain");
		if (userSelect) {
			return userSelect.title;
		} else if (info.title) {
			return info.title;
		} else if (info.tagName === "A") {
			// TODO(john): is this path used anymore?
			return info.textContent.trim(); // get text and strip whitespace
		} else if (info.tagName === "DIV") {
			const link = info.querySelector("a");
			if (link) {
				return link.title ? link.title : link.textContent.trim();
			} else {
				return null;
			}
		} else {
			return null;
		}
	} else {
		return path;
	}
}

function injectBackgroundApp() {
	let backgroundContainer = document.createElement("div");
	backgroundContainer.id = "sourcegraph-app-background";
	backgroundContainer.style.display = "none";
	document.body.appendChild(backgroundContainer);
	injectComponent(<Background />, backgroundContainer);
}

function injectBlobAnnotator() {
	if (!isGitHubURL()) return;

	const nTotalLines = getLinesOfCode();
	const {user, repo, rev, path, isDelta} = parseURL();
	const files = document.getElementsByClassName("file");

	for (let i = 0; i < files.length; ++i) {
		const file = files[i];
		const info = file.querySelector(".file-info");
		const blob = file.querySelector(".blob-wrapper");
		const actn = file.querySelector(".file-actions");
		const note = file.querySelector(".show-file-notes");
		const btng = actn ? actn.querySelector(".BtnGroup") : null;

		if (!blob) continue;

		const infoFilePath = getFileName(info, {isDelta, path});
		if (!infoFilePath) continue;

		const blobAnnotatorContainer = document.createElement("button");
		blobAnnotatorContainer.className = "btn btn-sm tooltipped tooltipped-n sourcegraph-app-annotator";
		blobAnnotatorContainer.style.display = "inline-block";
		blobAnnotatorContainer.style.verticalAlign = "middle";
		blobAnnotatorContainer.style.marginTop = info.tagName === 'A' ? "2px" : "0px";
		blobAnnotatorContainer.style.marginRight = info.tagName === 'A' && !file.querySelector(".show-outdated-button") ? "0px" : "5px";

		if (actn) {
			if (btng) {
				blobAnnotatorContainer.style.float = "none";
				btng.parentNode.insertBefore(blobAnnotatorContainer, btng);
			} else {
				if (note) {
					blobAnnotatorContainer.style.float = "none";
					note.parentNode.insertBefore(blobAnnotatorContainer, note.nextSibling);
				} else {
					blobAnnotatorContainer.style.float = "left";
					actn.appendChild(blobAnnotatorContainer);
				}
			}
		} else {
			if (info) {
				blobAnnotatorContainer.style.float = "right";
				info.parentNode.insertBefore(blobAnnotatorContainer, info.nextSibling);
			} else {
				blobAnnotatorContainer.style.float = "left";
				files.appendChild(blobAnnotatorContainer);
			}
		}

		if (nTotalLines >= isTooBigThreshold) {
			let iconLabelNode = document.createElement("span");
			let textLabelNode = document.createTextNode("Sourcegraph");

			blobAnnotatorContainer.appendChild(iconLabelNode);
			blobAnnotatorContainer.appendChild(textLabelNode);
			blobAnnotatorContainer.setAttribute("aria-label", readableGitHubRoute[getGitHubRoute()] + " too big!");
			blobAnnotatorContainer.setAttribute("disabled", true);

			injectComponent(<SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px", WebkitFilter: "grayscale(100%)"}} />, iconLabelNode);
		} else {
			injectComponent(<BlobAnnotator path={infoFilePath} blobElement={blob} infoElement={info} selfElement={blobAnnotatorContainer} />, blobAnnotatorContainer);
		}
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
	if (!document.getElementById('sourcegraph-app-bootstrap')) {
		injectBackgroundApp();
		injectBlobAnnotator();

		// Add invisible div to the page to indicate injection has completed.
		let el = document.createElement("div");
		el.id = "sourcegraph-app-bootstrap";
		el.style.display = "none";
		document.body.appendChild(el);
	}
}

window.addEventListener("load", () => {
	if (isSourcegraphURL()) {
		injectModules();
	} else if (isGitHubURL()) {
		chrome.runtime.sendMessage(null, {type: "getSessionToken"}, {}, (token) => {
			store.dispatch(Actions.setAccessToken(token));
			injectModules();
		});
	}
	chrome.runtime.sendMessage(null, {type: "getIdentity"}, {}, (identity) => {
		if (identity) EventLogger.updatePropsForUser(identity);
	});
});

document.addEventListener("keydown", (e) => {
	if (getGitHubRoute() !== "blob") return;
	if (e.target.tagName === "INPUT" ||
		e.target.tagName === "SELECT" ||
		e.target.tagName === "TEXTAREA") return;

	if (e.keyCode === 85) {
		const annButtons = document.getElementsByClassName("sourcegraph-app-annotator");
		if (annButtons.length === 1) {
			const annButtonA = annButtons[0].getElementsByTagName("A");
			if (annButtonA.length === 1 && annButtonA[0].href) {
				window.open(annButtonA[0].href, "_blank");
			}
		}
	}
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
