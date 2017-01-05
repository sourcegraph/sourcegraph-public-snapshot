import * as autobind from "autobind-decorator";
import * as React from "react";
import { browserHistory } from "react-router";

import { abs, getRouteParams } from "sourcegraph/app/routePatterns";
import { BlobStore } from "sourcegraph/blob/BlobStore";
import { BlobTitle } from "sourcegraph/blob/BlobTitle";
import { toggleCodeLens } from "sourcegraph/workbench/ConfigurationService";
import { PathSpec } from "sourcegraph/workbench/utils";

interface Props {
	pathspec: PathSpec;
}

@autobind
export class EditorTitle extends React.Component<Props, {}> {
	static childContextTypes: any = {
		router: React.PropTypes.object.isRequired,
	};
	toDispose: { remove(): void };

	componentWillUpdate(): void {
		this.toDispose = BlobStore.addListener(this.forceUpdate);
	}

	componentWillUnmount(): void {
		this.toDispose.remove();
	}

	getChildContext(): any {
		const router = browserHistory;
		router.setRouteLeaveHook = () => {
			throw "Cannot set route leave hook outside React Router hiearchy.";
		};
		router.isActive = () => {
			throw "Cannot access isActive outside React Router hiearchy.";
		};
		return {
			router: browserHistory,
		};
	}

	render(): JSX.Element {
		let {repo, rev, path} = this.props.pathspec;
		if (rev === "HEAD") {
			rev = null;
		}
		const params = getRouteParams(abs.blob, document.location.pathname);
		return <BlobTitle
			repo={repo}
			rev={rev}
			path={path}

			routeParams={params}
			toggleAuthors={toggleCodeLens}
			routes={[
				{ path: "/*/-/blob/*" },
			]}
			toast={BlobStore.toast}
			/>;
	}
}
