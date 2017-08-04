import * as React from "react";
import Resizable from "react-resizable-box";
import { ReactTree } from "sourcegraph/components/FileTree/ReactTree";
import { TreeHeader } from "sourcegraph/components/FileTree/TreeHeader";
import { TreeNode } from "sourcegraph/components/FileTree/util";

interface Props {
	parentRef: HTMLElement;
	treeData: TreeNode[];
	onChanged: (items: any[]) => void;
	onToggled?: (toggled: boolean) => void;
	toggled: boolean;
}

interface State {
	toggled: boolean;
}

export class TreeViewer extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {
			toggled: this.props.toggled,
		};

		document.querySelector("#navigation")!.addEventListener("click", (e: Event) => {
			e.preventDefault();
			this.toggleTreeViewer();
		});

		document.addEventListener("keydown", (event: any) => {
			if (event.altKey && event.code === "KeyT") {
				this.toggleTreeViewer();
			}
		});
	}

	onResize(_: () => void, __: string, element: HTMLElement): void {
		this.props.parentRef.style.width = element.style.width;
	}

	toggleTreeViewer(): void {
		if (this.props.onToggled) {
			this.props.onToggled(!this.state.toggled);
		}
		this.setState({
			...this.state, toggled: !this.state.toggled,
		});
	}

	render(): JSX.Element | null {
		if (!this.props.treeData) {
			return null;
		}
		if (!this.state.toggled) {
			return null;
		}
		return (
			<div style={{ display: "flex", height: "100%" }}>
				<Resizable onResize={this.onResize.bind(this)} className="item" width="100%" height="100%">
					<TreeHeader onClick={this.toggleTreeViewer.bind(this)} />
					<ReactTree onChanged={this.props.onChanged} plugins={["wholerow"]} core={{ multiple: false, worker: false, data: this.props.treeData }} />
				</Resizable>
			</div>
		);
	}
}
