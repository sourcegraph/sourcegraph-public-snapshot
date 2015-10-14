import React from "react";
import URI from "urijs";

import CodeFileContainer from "./CodeFileContainer";

// All data from window.location gets processed here and is then passed down
// to sub-components via props. Every time window.location changes, this
// component gets re-rendered. Sub-components should never access
// window.location by themselves.
class CodeFileRouter extends React.Component {
	constructor(props, context) {
		super(props, context);
	}

	componentDidMount() {
		window.addEventListener("popstate", this._locationChanged.bind(this));
	}

	componentWillUnmount() {
		window.removeEventListener("popstate", this._locationChanged.bind(this));
	}

	_locationChanged() {
		this.forceUpdate(); // this is necessary because the component uses external state (window.location)
	}

	render() {
		let uri = URI.parse(window.location.href);
		let pathParts = uri.path.substr(1).split("/.");

		let keys = [];
		let vars = URI.parseQuery(uri.query);

		let repoParts = pathParts[0].split("@");
		vars["repo"] = repoParts[0];
		vars["rev"] = repoParts[1];

		pathParts.slice(1).forEach((part) => {
			let p = part.indexOf("/");
			let key = part.substr(0, p);
			keys.push(key);
			vars[key] = part.substr(p + 1);
		});

		if (vars["def"] !== undefined) {
			return (
				<CodeFileContainer
					repo={vars["repo"]}
					rev={vars["rev"]}
					unitType={keys[0]}
					unit={vars[keys[0]]}
					def={vars["def"]}
					example={vars["examples"] && parseInt(vars["examples"], 10)} />
			);
		}

		return (
			<CodeFileContainer
				repo={vars["repo"]}
				rev={vars["rev"]}
				tree={vars["tree"]}
				token={vars["token"] && parseInt(vars["token"], 10)}
				startline={vars["startline"] && parseInt(vars["startline"], 10)}
				endline={vars["endline"] && parseInt(vars["endline"], 10)} />
		);
	}
}

export default CodeFileRouter;
