import React from "react";
import ReactDOM from "react-dom";
import requireComponent from "sourcegraph/init/requireComponent";
import resetStores from "sourcegraph/init/resetStores";

let logTimings = false;
if (typeof window !== "undefined") {
	logTimings = window.localStorage["log-timings"] === "true";

	window.enableTimingsLog = function() {
		window.localStorage["log-timings"] = "true";
		console.log("Timings log enabled.");
	};

	window.disableTimingsLog = function() {
		Reflect.deleteProperty(window.localStorage, "log-timings");
		console.log("Timings log disabled.");
	};
}

if (typeof document !== "undefined") {
	if (window.__StoreData) {
		resetStores(window.__StoreData);
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
	setTimeout(() => {
		const label = props && props.component ? props.component.name : Component.name;
		if (logTimings) console.time(`render ${label}`);
		ReactDOM.render(<Component {...props} />, el);
		if (logTimings) console.timeEnd(`render ${label}`);
	}, 0);
}
