import React from "react";

import Component from "../Component";

import ReactDOM from "react-dom";
import ReactCSSTransitionGroup from "react-addons-css-transition-group";


// Note: for this tooltip to appear correctly, its parent element must have position
// style attribute set to either "relative" or "aboslute".
class Tooltip extends Component {
	constructor(props) {
		super(props);
		this._onHover = this._onHover.bind(this);
		this._onUnhover = this._onUnhover.bind(this);
	}

	componentDidMount() {
		let el = ReactDOM.findDOMNode(this).parentElement;
		el.addEventListener("mouseenter", this._onHover);
		el.addEventListener("mouseleave", this._onUnhover);
	}

	componentWillUnmount() {
		let el = ReactDOM.findDOMNode(this).parentElement;
		el.removeEventListener("mouseenter", this._onUnhover);
		el.removeEventListener("mouseleave", this._onUnhover);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onHover() {
		if (!this.state.show) this.setState({show: true});
	}

	_onUnhover() {
		if (this.state.show) this.setState({show: false});
	}

	render() {
		let style = {
			display: this.state.show ? "block" : "none",
		};

		return (
			<ReactCSSTransitionGroup transitionName="fade" transitionEnterTimeout={600} transitionLeaveTimeout={1} style={style}>
				{this.state.show &&
					<div key="tooltip-sg" className="tooltip-sg">
						<div key="tooltip-inner" className="tooltip-inner">
							{this.state.children}
						</div>
					</div>
				}
			</ReactCSSTransitionGroup>
		);
	}
}

export default Tooltip;
