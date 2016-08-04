// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/flexContainer.css";


class FlexContainer extends React.Component<any, any> {
	static propTypes = {
		direction: React.PropTypes.string, // left_right, right_left, top_bottom, bottom_top
		wrap: React.PropTypes.bool,
		justify: React.PropTypes.string, // start, end, center, between, around
		items: React.PropTypes.string, // start, end, center, baseline, stretch
		content: React.PropTypes.string, // start, end, center, between, around, stretch
		className: React.PropTypes.string,
		children: React.PropTypes.any,
	};

	static defaultProps = {
		direction: "left_right",
		wrap: false,
		justify: "start",
		items: "stretch",
		content: "stretch",
	};

	render(): JSX.Element | null {
		const {direction, wrap, justify, items, content, className, children} = this.props;
		return (
			<div styleName={`flex ${direction} justify_${justify} items_${items} content_${content} ${wrap ? "wrap" : "nowrap"}`} className={className}>
				{children}
			</div>
		);
	}
}

export default CSSModules(FlexContainer, styles, {allowMultiple: true});
