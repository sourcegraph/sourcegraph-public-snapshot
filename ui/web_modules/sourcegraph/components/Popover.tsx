import * as classNames from "classnames";
import * as React from "react";

import {Component, EventListener} from "sourcegraph/Component";
import * as styles from "sourcegraph/components/styles/popover.css";

interface Props {
	left?: boolean; // position popover content to the left (default: right)
	popoverClassName?: string;
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

	_onClick(e: MouseEvent & {target: Node}): void {
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
			<div className={styles.container} ref={this.setContainer}>
				{this.props.children && this.props.children[0]}
				{this.state.visible &&
					<div ref={this.setContent} className={classNames(this.props.left ? styles.popover_left : styles.popover_right, this.props.popoverClassName)}>
						{this.props.children[1]}
					</div>}
				<EventListener target={global.document} event="click" callback={this._onClick} />
			</div>
		);
	}
}
