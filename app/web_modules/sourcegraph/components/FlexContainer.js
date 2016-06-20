// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/flexContainer.css";


class FlexContainer extends React.Component {
	static propTypes = {
		direction: React.PropTypes.string, // left-right, right-left, top-bottom, bottom-top
		wrap: React.PropTypes.bool,
		justify: React.PropTypes.string, // start, end, center, between, around
		items: React.PropTypes.string, // start, end, center, baseline, stretch
		content: React.PropTypes.string, // start, end, center, between, around, stretch
		className: React.PropTypes.string,
		children: React.PropTypes.any,
	};

	static defaultProps = {
		direction: "left-right",
		wrap: false,
		justify: "start",
		items: "stretch",
		content: "stretch",
	};

	render() {
		const {direction, wrap, justify, items, content, className, children} = this.props;
		return (
			<div styleName={`flex ${direction} justify-${justify} items-${items} content-${content} ${wrap ? "wrap" : "nowrap"}`} className={className}>
				{children}
			</div>
		);
	}
}

export default CSSModules(FlexContainer, styles, {allowMultiple: true});
