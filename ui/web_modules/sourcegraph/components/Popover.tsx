import { css } from "glamor";
import * as React from "react";

import { EventListener } from "sourcegraph/Component";
import * as colors from "sourcegraph/components/utils/colors";

interface Props {
	left?: boolean; // position popover content to the left (default: right)
	pointer?: boolean;
	popoverClassName?: string;
	popoverStyle?: any;
	children?: React.ReactNode;
}

type State = {
	visible: boolean;
};

export class Popover extends React.Component<Props, State> {
	content: HTMLElement;
	container: HTMLElement;

	constructor(props: Props) {
		super(props);

		if (!(props.children instanceof Array) || props.children.length !== 2) {
			throw new Error("Popover must be constructed with exactly two children.");
			// TODO(chexee): make this accomodate multiple lengths!
		}

		this.state = {
			visible: false,
		};
		this._onClick = this._onClick.bind(this);
		this.setContainer = this.setContainer.bind(this);
		this.setContent = this.setContent.bind(this);
	}

	_onClick(e: MouseEvent & { target: Node }): void {
		let container = this.container;
		let content = this.content;
		if (container && container.contains(e.target)) {
			// Toggle popover visibility when clicking on container or anywhere else
			this.setState({
				visible: !this.state.visible,
			} as State);
		} else if (content && !content.contains(e.target)) {
			// Dismiss popover when clicking on anything else but content.
			this.setState({
				visible: false,
			} as State);
		}
	}

	setContent(ref: HTMLElement): void {
		this.content = ref;
	}
	setContainer(ref: HTMLElement): void {
		this.container = ref;
	}

	render(): JSX.Element | null {
		if (!this.props.children) { return null; }
		return (
			<div style={{ position: "relative", cursor: "pointer" }} ref={this.setContainer}>
				{this.props.children && this.props.children[0]}
				{this.state.visible &&
					<div ref={this.setContent}
						{...css(
							{
								borderRadius: 3,
								minWidth: 100,
								cursor: "initial",
								position: "absolute",
								top: "97%",
								left: this.props.left ? "" : -8,
								right: this.props.left ? -8 : "",
								zIndex: 100,
							},
							this.props.pointer ? {
								":before": {
									content: `""`,
									backgroundColor: "white",
									borderLeft: `1px ${colors.blueGrayL1(0.2)} solid`,
									borderTop: `1px ${colors.blueGrayL1(0.2)} solid`,
									display: "block",
									height: 8,
									position: "absolute",
									right: 16,
									top: -4,
									transform: "rotate(45deg) skew(-10deg, -10deg)",
									width: 8,
									zIndex: 101,
								}
							} : {},
							this.props.popoverStyle,
						) }>
						{this.props.children[1]}
					</div>
				}
				<EventListener target={global.document} event="click" callback={this._onClick} />
			</div>
		);
	}
}
