// tslint:disable

import * as React from "react";

import {Component} from "sourcegraph/Component";

export class Collapsible extends Component<Props, any> {
	constructor(props) {
		super(props);

		if (props.children.length !== 2) {
			throw new Error("Collapsible must be constructed with exactly two children.");
		}

		this.state = {
			shown: !props.collapsed,
		};
		this._onClick = this._onClick.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onClick() {
		const isShown = !this.state.shown;
		this.setState({
			shown: isShown,
		}, () => this.state.onToggle && this.state.onToggle(isShown));
	}

	render(): JSX.Element | null {
		return (
			<div>
				<div onClick={this._onClick} style={{cursor: "pointer"}}>{this.state.children[0]}</div>
				{this.state.shown && this.state.children[1]}
			</div>
		);
	}
}

type Props = {
	collapsed?: boolean,
	onToggle?: () => void,
};
