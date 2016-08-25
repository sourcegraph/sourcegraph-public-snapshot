// tslint:disable: typedef ordered-imports

import * as React from "react";
import {PlainRoute, RouterState} from "react-router";
import {Header} from "sourcegraph/components/Header";
import {defaultFetch} from "sourcegraph/util/xhr";

// The "/-/golang" route acts as a redirect. withChannelListener receives information through the websocket
// and, if there are no errors, replaces the path with "-/golang". Once the user enters the "-/golang" path
// (i.e. onEnter), golang.js figures out the correct path the user wants , whether a tree page or a (def)info
// page.

interface Props {
	location: any;
}

function GoLookup(props: Props, context) {
	return (
		<Header title={`Go code not found: '${props.location.query.def}'`}
			subtitle="" />
	);
}

export const route: PlainRoute = {
	path: "-/golang",
	onEnter: (nextRouterState: RouterState, replace: Function, callback: Function) => {
		let repo = nextRouterState.location.query["repo"];
		let pkg = nextRouterState.location.query["pkg"];
		let def = nextRouterState.location.query["def"];
		let editorType = nextRouterState.location.query["editor_type"];
		if (/\.(com|org|net|in)\//.test(repo)) {
			repo = repo.split("/").slice(0, 3).join("/");
		}
		if (def) {
			defaultFetch(`/.api/resolve-custom-import/info?def=${def}&pkg=${pkg}&repo=${repo}`)
				.then((resp) => resp.json())
				.then((data) => {
					// TODO(matt): remove once sourcegraph.com resolving bug is fixed
					let path = data.Path;
					if (path.startsWith("/github.com/sourcegraph/sourcegraph")) {
						path = path.replace("GoPackage/github.com", "GoPackage/sourcegraph.com");
					}
					replace(Object.assign({}, nextRouterState.location, {
						pathname: path,
						query: {
							utm_source: "sourcegraph-editor",
							editor_type: editorType,
						},
					}));
					callback();
				});
		} else {
			defaultFetch(`/.api/resolve-custom-import/tree?repo=${repo}&pkg=${pkg}`)
				.then((resp) => resp.json())
				.then((data) => {
					replace(Object.assign({}, nextRouterState.location, {
						pathname: data.Path,
						query: {
							utm_source: "sourcegraph-editor",
							editor_type: editorType,
						},
					}));
					callback();
				});
		}
	},
	components: {
		main: GoLookup,
	},
};
