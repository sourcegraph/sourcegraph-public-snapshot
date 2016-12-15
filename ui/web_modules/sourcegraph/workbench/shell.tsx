import * as autobind from "autobind-decorator";
import * as React from "react";
import { IEditorService } from "vs/platform/editor/common/editor";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { URIUtils } from "sourcegraph/core/uri";

interface Props {
	repo: string;
	rev: string | null;
	path: string;
};

interface State {};

// Shell loads the workbench and calls init on it. It transmits data from the
// React UI layer into the Workbench interface. It is primarily controlled by
// React Router. It uses code splitting to minimize bundle size.
export class Shell extends React.Component<Props, State> {
	workbench: Workbench;
	services: ServiceCollection;

	private mounted: boolean = false;

	@autobind
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
		});
	}

	componentWillReceiveProps(nextProps: Props): void {
		// if (!this.mounted || !this.workbench) {
		// 	return;
		// }
		// const resource = URIUtils.pathInRepo(this.props.repo, this.props.rev, this.props.path);
		// const editorService = this.services.get(IEditorService) as IEditorService;
		// editorService.openEditor({resource});
		// TODO scroll to the right position.
	}

	render(): JSX.Element {
		return <div className={"vs-dark"} style={{height: "100%"}} ref={this.domRef}></div>;
	}

}
