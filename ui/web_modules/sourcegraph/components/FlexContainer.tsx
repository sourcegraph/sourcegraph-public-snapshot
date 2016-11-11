import * as classNames from "classnames";
import * as React from "react";

import * as styles from "sourcegraph/components/styles/flexContainer.css";

interface Props {
	direction?: string; // left_right, right_left, top_bottom, bottom_top
	wrap?: boolean;
	justify?: string; // start, end, center, between, around
	items?: string; // start, end, center, baseline, stretch
	content?: string; // start, end, center, between, stretch
	className?: string;
	children?: any;
	style?: Object;
}

export class FlexContainer extends React.Component<Props, {}> {
	static defaultProps: Props = {
		direction: "left_right",
		wrap: false,
		justify: "start",
		items: "stretch",
		content: "stretch",
	};

	render(): JSX.Element | null {
		const {direction = "left_right", wrap, justify = "start", items = "stretch", content = "stretch", className, children, style} = this.props;

		return (
			<div className={classNames(styles.flex, directionClasses[direction], justifyClasses[justify], itemsClasses[items], contentClasses[content], wrap ? styles.wrap : styles.nowrap, className)} style={style}>
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
