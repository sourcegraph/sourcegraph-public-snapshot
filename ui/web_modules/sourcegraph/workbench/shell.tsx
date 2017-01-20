import * as autobind from "autobind-decorator";
import * as debounce from "lodash/debounce";
import * as isEqual from "lodash/isEqual";
import * as React from "react";

import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { BlobRouteProps, Router } from "sourcegraph/app/router";
import { URIUtils } from "sourcegraph/core/uri";
import { registerEditorCallbacks, syncEditorWithRouterProps } from "sourcegraph/editor/config";
import { init } from "sourcegraph/workbench/main";

// WorkbenchShell loads the workbench and calls init on it. It is a pure container and transmits no data from the
// React UI layer into the Workbench interface. Synchronization of URL <-> workbench is handled by
// adding a listener to the "sourcegraph/app/router" package, and by pushing updates to the singleton
// router from that package.
@autobind
export class WorkbenchShell extends React.Component<BlobRouteProps, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };
	workbench: Workbench;
	services: ServiceCollection;
	listener: number;

	domRef(domElement: HTMLDivElement): void {
		if (!domElement) {
			if (this.workbench) {
				this.workbench.dispose();
			}
			return;
		}

		const {repo, rev, path} = this.props;
		const resource = URIUtils.pathInRepo(repo, rev, path);
		[this.workbench, this.services] = init(domElement, resource);
		registerEditorCallbacks(this.context.router);

		this.layout();
		syncEditorWithRouterProps(this.props);
	}

	componentWillMount(): void {
		window.onresize = debounce(this.layout, 50);
	}

	componentWillUnmount(): void {
		window.onresize = () => void (0);
	}

	componentWillReceiveProps(nextProps: BlobRouteProps): void {
		if (!isEqual(nextProps, this.props)) {
			syncEditorWithRouterProps(nextProps);
		}
	}

	layout(): void {
		if (!this.workbench) {
			return;
		}
		if (window.innerWidth <= 768) {
			// Mobile device, width less than 768px.
			this.workbench.setSideBarHidden(true);
		} else {
			this.workbench.setSideBarHidden(false);
		}
		this.workbench.layout();
	}

	render(): JSX.Element {
		this.layout();
		return <div className="vs-dark" style={{
			height: "100%",
			flex: "1 1 100%",
		}} ref={this.domRef}></div>;
	}

}
