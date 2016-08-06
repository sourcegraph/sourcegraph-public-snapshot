// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "sourcegraph/components/styles/flexContainer.css";


class FlexContainer extends React.Component<any, any> {
	static propTypes = {
		direction: React.PropTypes.string, // left_right, right_left, top_bottom, bottom_top
		wrap: React.PropTypes.bool,
		justify: React.PropTypes.string, // start, end, center, between, around
		items: React.PropTypes.string, // start, end, center, baseline, stretch
		content: React.PropTypes.string, // start, end, center, between, stretch
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
			<div className={`${styles.flex} ${directionClasses[direction]} ${justifyClasses[justify]} ${itemsClasses[items]} ${contentClasses[content]} ${wrap ? styles.wrap : styles.nowrap} ${className}`}>
				{children}
			</div>
		);
	}
}

const directionClasses = {
	"left_right": styles.left_right,
	"right_left": styles.right_left,
	"top_bottom": styles.top_bottom,
	"bottom_top": styles.bottom_top,
};

const justifyClasses = {
	"start": styles.justify_start,
	"end": styles.justify_end,
	"center": styles.justify_center,
	"between": styles.justify_between,
	"around": styles.justify_around,
};

const itemsClasses = {
	"start": styles.items_start,
	"end": styles.items_end,
	"center": styles.items_center,
	"baseline": styles.items_baseline,
	"stretch": styles.items_stretch,
};

const contentClasses = {
	"start": styles.content_start,
	"end": styles.content_end,
	"center": styles.content_center,
	"between": styles.content_between,
	"stretch": styles.content_stretch,
};

export default CSSModules(FlexContainer, styles, {allowMultiple: true});
