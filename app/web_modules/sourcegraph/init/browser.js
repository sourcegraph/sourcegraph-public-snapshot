import React from "react";
import ReactDOM from "react-dom";
import requireComponent from "sourcegraph/init/requireComponent";

import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import RepoStore from "sourcegraph/repo/RepoStore";

if (typeof document !== "undefined") {
	if (window.__StoreData) {
		RepoStore.reset(window.__StoreData.RepoStore || {});
		BlobStore.reset(window.__StoreData.BlobStore || {});
		DefStore.reset(window.__StoreData.DefStore || {});
	}

	let els = document.querySelectorAll("[data-react]");
	for (let i = 0; i < els.length; i++) {
		let el = els[i];
		const Component = requireComponent(el.dataset.react);
		let props = el.dataset.props ? JSON.parse(el.dataset.props) : null;
		if (props && props.component) {
			props.component = requireComponent(props.component);
		}
		render(Component, props, el);
	}
}

function render(Component, props, el) {
	setTimeout(() => ReactDOM.render(<Component {...props} />, el), 0);
}
