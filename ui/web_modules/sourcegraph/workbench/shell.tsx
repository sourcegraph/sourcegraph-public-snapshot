import * as autobind from "autobind-decorator";
import * as debounce from "lodash/debounce";
import * as React from "react";

import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { BlobRouteProps, Router, getBlobPropsFromRouter } from "sourcegraph/app/router";
import { URIUtils } from "sourcegraph/core/uri";
import { registerEditorCallbacks, syncEditorWithRouterProps, unmountWorkbench } from "sourcegraph/editor/config";

interface Props extends BlobRouteProps { }

// WorkbenchShell loads the workbench and calls init on it. It is a pure container and transmits no data from the
// React UI layer into the Workbench interface. Synchronization of URL <-> workbench is handled by
// adding a listener to the "sourcegraph/app/router" package, and by pushing updates to the singleton
// router from that package.
@autobind
export class WorkbenchShell extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };
	workbench: Workbench;
	services: ServiceCollection;
	listener: number;

	private mounted: boolean = false;

	domRef(domElement: HTMLDivElement): void {
		if (!domElement) {
			this.mounted = false;
			if (this.workbench) {
				this.workbench.dispose();
			}
			return;
		}
		this.mounted = true;

		// TODO(john): I don't think this is properly code-splitting, and it certainly won't
		// if we're already importing a bunch of vscode paths from config.tsx. Reconsider.
		require(["sourcegraph/workbench/main"], ({init}) => {
			if (!this.mounted) {
				// component unmounted before require finished.
				return;
			}
			const blobProps = getBlobPropsFromRouter(this.context.router);
			const {repo, rev, path} = blobProps;
			const resource = URIUtils.pathInRepo(repo, rev, path);
			[this.workbench, this.services] = init(domElement, resource);
			registerEditorCallbacks(this.context.router);

			this.layout();
			syncEditorWithRouterProps(blobProps);
		});
	}

	componentWillMount(): void {
		window.onresize = debounce(this.layout, 50);
	}

	componentWillUnmount(): void {
		window.onresize = () => void (0);
		unmountWorkbench();
	}

	componentWillReceiveProps(nextProps: Props): void {
		syncEditorWithRouterProps(nextProps);
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
