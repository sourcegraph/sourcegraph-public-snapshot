import * as autobind from "autobind-decorator";
import * as React from "react";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { URIUtils } from "sourcegraph/core/uri";
import { updateEditor } from "sourcegraph/editor/config";

interface Props {
	repo: string;
	rev: string | null;
	path: string;
};

interface State { };

// Shell loads the workbench and calls init on it. It transmits data from the
// React UI layer into the Workbench interface. It is primarily controlled by
// React Router. It uses code splitting to minimize bundle size.
@autobind
export class Shell extends React.Component<Props, State> {
	workbench: Workbench;
	services: ServiceCollection;

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
			const workspace = URIUtils.pathInRepo(this.props.repo, this.props.rev, this.props.path);
			[this.workbench, this.services] = init(domElement, workspace);
			this.layout();
		});
	}

	componentWillMount(): void {
		window.onresize = this.layout;
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

	componentWillUnmount(): void {
		window.onresize = () => void (0);
	}

	componentWillReceiveProps(nextProps: Props): void {
		if (!this.mounted || !this.workbench) {
			return;
		}
		const resource = URIUtils.pathInRepo(nextProps.repo, nextProps.rev, nextProps.path);
		updateEditor(this.workbench.getEditorPart(), resource, this.services);
	}

	render(): JSX.Element {
		return <div className="vs-dark" style={{ height: "100%" }} ref={this.domRef}></div>;
	}

}
