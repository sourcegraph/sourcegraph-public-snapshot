import * as autobind from "autobind-decorator";
import * as React from "react";

import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { getBlobPropsFromRouter } from "sourcegraph/app/router";
import { URIUtils } from "sourcegraph/core/uri";
import { syncEditorWithRouter } from "sourcegraph/editor/config";

// WorkbenchShell loads the workbench and calls init on it. It is a pure container and transmits no data from the
// React UI layer into the Workbench interface. Synchronization of URL <-> workbench is handled by
// adding a listener to the "sourcegraph/app/router" package, and by pushing updates to the singleton
// router from that package.
@autobind
export class WorkbenchShell extends React.Component<{}, {}> {
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
		require(["sourcegraph/workbench/main"], ({init}) => {
			if (!this.mounted) {
				// component unmounted before require finished.
				return;
			}
			const {repo, rev, path} = getBlobPropsFromRouter();
			const resource = URIUtils.pathInRepo(repo, rev, path);
			[this.workbench, this.services] = init(domElement, resource);
			this.layout();
		});
	}

	componentWillMount(): void {
		window.onresize = this.layout;
		window.addEventListener("popstate", this.synchronizeEditor);
	}

	componentWillUnmount(): void {
		window.onresize = () => void (0);
		window.removeEventListener("popstate", this.synchronizeEditor);
	}

	synchronizeEditor(): void {
		// Provides browser forward & back handling.
		syncEditorWithRouter(this.services);
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
		return <div className="vs-dark" style={{ height: "100%" }} ref={this.domRef}></div>;
	}

}
