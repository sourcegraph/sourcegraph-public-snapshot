import "sourcegraph/components/FileTree/jstree.css";
import "sourcegraph/components/FileTree/jstreeDark.css";

import * as React from "react";
import * as ReactDOM from "react-dom";

import * as $ from "jquery";
import "jstree";

interface Props {
	core: {
		data: any[],
		worker: boolean,
		multiple: boolean,
	};
	plugins?: string[];
	onChanged: (items: any[]) => void;
}

export class ReactTree extends React.Component<Props, {}> {
	componentDidMount(): void {
		($(ReactDOM.findDOMNode(this))
			.on("changed.jstree", (_: any, data: any): any => {
				if (this.props.onChanged) {
					this.props.onChanged(data.selected.map(
						item => data.instance.get_node(item),
					));
				}
			}) as any)
			.jstree({
				core: this.props.core,
				plugins: this.props.plugins,
			});
	}

	render(): JSX.Element | null {
		return <div />;
	}
}
