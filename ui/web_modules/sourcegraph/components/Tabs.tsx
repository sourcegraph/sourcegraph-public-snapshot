import * as classNames from "classnames";
import * as React from "react";

import * as styles from "sourcegraph/components/styles/tabs.css";

interface Props {
	direction?: string; // vertical, horizontal
	color?: string; // blue, purple
	size?: string; // small, large, default
	children?: any;
	className?: string;
	style?: any;
}

export class Tabs extends React.Component<Props, {}> {
	static defaultProps: Props = {
		size: "default",
		direction: "horizontal",
		color: "blue",
	};

	_childrenWithProps(): JSX.Element[] {
		return React.Children.map(this.props.children, (child: React.ReactElement<any>) => {
			if (child.props.tabItem) {
				return React.cloneElement(child, {
					direction: this.props.direction,
					color: this.props.color,
					size: this.props.size,
				});
			}
			return React.cloneElement(child, {
				className: this.props.direction === "vertical" ? styles.item_vertical : styles.item_horizontal,
			});
		});
	}

	render(): JSX.Element {
		const {direction, className, style} = this.props;
		return <div className={classNames(styles.container, direction === "vertical" ? styles.vertical : styles.horizontal, className)} style={style}>{this._childrenWithProps()}</div>;
	}
}
