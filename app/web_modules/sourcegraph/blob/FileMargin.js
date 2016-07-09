// @flow weak

import React from "react";
import ReactDOM from "react-dom";

if (typeof document !== "undefined" && document.body.style.setProperty) {
	// Compute code line height. It's not always the `codeLineHeight`
	// value, when full-page zoom is being used, for example. This is
	// necessary to properly align the boxes to the code on 90%, 100%,
	// etc., full-page zoom levels.
	let el = document.createElement("div");
	el.innerText = "a";
	document.body.appendChild(el);
	document.body.removeChild(el);
}


export default class FileMargin extends React.Component {

	constructor(props) {
		super(props);
		this.state = {
			offsetFromTop: 0,
		};
	}

	componentWillReceiveProps() {
		this.setState({offsetFromTop: this.getOffsetFromTop()});
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

		let i = -1;
		return (
			<div {...passthroughProps} style={{position: "relative"}}>
				{React.Children.map(this.props.children, (child) => {
					i++;
					return (
						<div key={i} style={{position: "absolute", top: `${this.state.offsetFromTop}px`}}>
							{child}
						</div>
					);
				})}
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
};
