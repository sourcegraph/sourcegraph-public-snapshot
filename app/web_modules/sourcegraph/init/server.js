import React from "react";
import ReactDOMServer from "react-dom/server";
import requireComponent from "sourcegraph/init/requireComponent";
import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import ReactMarkupChecksum from "react/lib/ReactMarkupChecksum";

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

		// HACK: We manually set the data-react-id on the <div
		// class="blob-scroller" of the ServerBlob React component so it
		// matches what the client expects. This is because we render the
		// Blob innerHTML in Go, outside of the normal React system, and
		// it automatically appends a data-reactid to each element. We
		// want it to use ours.
		//
		// So, remove theirs.
		//
		// We then need to update the React checksum to make the client
		// think everything's OK.
		let htmlStr = ReactDOMServer.renderToString(<Component {...arg.Props} />);
		htmlStr = htmlStr.replace(/ data-remove-second-reactid="yes" data-reactid="[^"]+"/, "");
		htmlStr = htmlStr.replace(/ data-react-checksum="[^"]+"/, "");
		htmlStr = ReactMarkupChecksum.addChecksumToMarkup(htmlStr);

		callback(htmlStr);
	};
}
