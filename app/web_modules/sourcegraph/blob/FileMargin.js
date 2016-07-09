// @flow

import React from "react";
import ReactDOM from "react-dom";

export default class FileMargin extends React.Component {

	componentDidUpdate() {
		if (this.refs.content) {
			const lineOffsetFromTop = this.getOffsetFromTop();
			const isNearBottom = lineOffsetFromTop > this.refs.content.parentNode.clientHeight - this.refs.content.clientHeight;

			this.refs.content.style.top = isNearBottom ? "" : `${lineOffsetFromTop}px`;
			this.refs.content.style.bottom = isNearBottom ? "0px" : "";
		}
	}

	getOffsetFromTop() {
		if (this.props.selectionStartLine) {
			return ReactDOM.findDOMNode(this.props.selectionStartLine).offsetTop;
		}
		return 0;
	}

	render() {
		let passthroughProps = {...this.props};
		delete passthroughProps.children;
		delete passthroughProps.lineFromByte;

		return (
			<div {...passthroughProps} style={{position: "relative"}}>
				{React.Children.map(this.props.children, (child, i) => (
					<div key={i} ref="content" style={{position: "absolute"}}>{child}</div>
				))}
			</div>
		);
	}
}
FileMargin.propTypes = {
	children: React.PropTypes.oneOfType([
		React.PropTypes.arrayOf(React.PropTypes.element),
		React.PropTypes.element,
	]),

	lineFromByte: React.PropTypes.func,
	selectionStartLine: React.PropTypes.any,
	startByte: React.PropTypes.number,
};
