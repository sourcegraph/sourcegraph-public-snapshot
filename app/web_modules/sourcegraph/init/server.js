import React from "react";
import ReactDOMServer from "react-dom/server";
import requireComponent from "sourcegraph/init/requireComponent";
import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";

// Only run on server (not in browser).
if (typeof document === "undefined") {
	// main is called from the Go reactbridge package to render React
	// components on the server (to increase the initial load speed).
	global.main = (arg, callback) => {
		BlobStore.reset((arg.Stores && arg.Stores.BlobStore) || {});
		DefStore.reset((arg.Stores && arg.Stores.DefStore) || {});

		const Component = requireComponent(arg.ComponentModule);
		if (arg.Props && arg.Props.component) {
			arg.Props.component = requireComponent(arg.Props.component);
		}

		callback(ReactDOMServer.renderToString(<Component {...arg.Props} />));
	};
}
