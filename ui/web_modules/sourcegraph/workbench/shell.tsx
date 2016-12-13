import * as autobind from "autobind-decorator";
import * as React from "react";
import { Workbench } from "vs/workbench/electron-browser/workbench";

interface Props {};

interface State {};

@autobind
export class Shell extends React.Component<Props, State> {
	workbench: Workbench;
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
			this.workbench = init(domElement);
		});
	}

	render(): JSX.Element {
		return <div style={{height: "100%"}} ref={this.domRef}></div>;
	}

}
