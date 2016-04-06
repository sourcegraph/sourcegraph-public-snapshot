import React from "react";
import ReactDOMServer from "react-dom/server";
import requireComponent from "sourcegraph/init/requireComponent";
import resetStores from "sourcegraph/init/resetStores";
import split from "split";
import {disableFetch} from "sourcegraph/util/xhr";

// handle is called from Go to render the page's contents.
const handle = (arg, callback) => {
	if (arg.Stores) {
		resetStores(arg.Stores);
	}

	const Component = requireComponent(arg.ComponentModule);
	if (arg.Props && arg.Props.component) {
		arg.Props.component = requireComponent(arg.Props.component);
	}

	let htmlStr = ReactDOMServer.renderToString(<Component {...arg.Props} />);

	callback(htmlStr);
};

// jsserver: listens on stdin for lines of JSON sent by the app/internal/ui Go package.
if (typeof global !== "undefined" && global.process && global.process.env.JSSERVER) {
	global.process.stdout.write("\"ready\"\n");
	console.log = console.error;
	disableFetch();

	global.process.stdin.pipe(split())
		.on("data", (line) => {
			if (line === "") return;
			handle(JSON.parse(line), (data) => {
				global.process.stdout.write(JSON.stringify(data));
				global.process.stdout.write("\n");
			});
		})
		.on("error", (err) => {
			console.error("jsserver: error reading line from stdin:", err);
		});
}
