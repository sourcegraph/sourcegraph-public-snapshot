import React from "react";
import ReactDOM from "react-dom";
import Component from "sourcegraph/Component";

export default class FileMargin extends Component {
	componentDidMount() {
		if (super.componentDidMount) super.componentDidMount();
		this._positionCards();
	}

	_positionCards() {
		if (!this.state.getOffsetTopForByte) return;
		// if (!this.state.getOffsetTopForByte) throw new Error("getOffsetTopForByte not provided (should have been provided via CodeFileContainer refs)");

		let $el = ReactDOM.findDOMNode(this);
		let $contentView = $el.parentNode;
		this._refs.forEach((c) => {
			let $c = ReactDOM.findDOMNode(c);
			$c.style.position = "absolute";
			$c.style.top = `${this.state.getOffsetTopForByte(c.props.def.ByteStartPosition) - $contentView.offsetTop}px`;
		});
	}

	reconcileState(state, props) {
		if (state.children !== props.children) {
			this._refs = [];
			state.children = React.Children.map(props.children, (c) =>
				c ? React.cloneElement(c, {
					ref: (e) => e ? this._refs.push(e) : null,
				}) : null, this);
		}
		if (state.getOffsetTopForByte !== props.getOffsetTopForByte) {
			state.getOffsetTopForByte = props.getOffsetTopForByte;
			setTimeout(this._positionCards.bind(this), 0);
		}
	}

	render() {
		return (
			<div className="sidebar file-sidebar" style={{position: "relative"}}>
				{this.state.children}
			</div>
		);
	}
}
FileMargin.propTypes = {
	children: React.PropTypes.oneOfType([
		React.PropTypes.arrayOf(React.PropTypes.element),
		React.PropTypes.element,
	]).isRequired,
	getOffsetTopForByte: React.PropTypes.func,
};
