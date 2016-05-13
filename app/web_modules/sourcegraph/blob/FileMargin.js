// @flow weak

import React from "react";
// $FlowHack
import {codeLineHeight, firstCodeLineTopPadding} from "sourcegraph/blob/styles/Blob.css";

let computedCodeLineHeight = codeLineHeight;

if (typeof document !== "undefined" && document.body.style.setProperty) {
	// Compute code line height. It's not always the `codeLineHeight`
	// value, when full-page zoom is being used, for example. This is
	// necessary to properly align the boxes to the code on 90%, 100%,
	// etc., full-page zoom levels.
	let el = document.createElement("div");
	el.style.lineHeight = codeLineHeight;
	el.innerText = "a";
	document.body.appendChild(el);
	computedCodeLineHeight = `${el.getBoundingClientRect().height}px`;
	document.body.removeChild(el);
}


export default class FileMargin extends React.Component {
	state = {codeLineHeight: codeLineHeight};

	componentDidMount() {
		// Initially render with the base line height, and then we'll later
		// update with the computed line height to account for full-page zoom
		// fractional pixel heights. To test this, reload the page at various
		// full-page zoom levels.
		if (this.state.codeLineHeight !== computedCodeLineHeight) {
			setTimeout(() => {
				this.setState({codeLineHeight: computedCodeLineHeight});
			});
		}
	}

	// _childOffsetTop is the CSS height expression from the top of the container that the
	// child should be offset.
	_childOffsetTop(i) {
		if (!this.props.lineFromByte) return null;
		const child = React.Children.toArray(this.props.children)[i];
		const lineIndex = this.props.lineFromByte(child.props.byte) - 1; // 0-indexed so line 1 is at 0px
		return `calc(${lineIndex} * ${this.state.codeLineHeight} + ${firstCodeLineTopPadding})`;
	}

	render() {
		let passthroughProps = {...this.props};
		delete passthroughProps.children;
		delete passthroughProps.lineFromByte;

		let i = -1;
		return (
			<div {...passthroughProps}>
				{React.Children.map(this.props.children, (child) => {
					i++;
					return (
						<div key={i} style={{marginTop: this._childOffsetTop(i)}}>
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
};
