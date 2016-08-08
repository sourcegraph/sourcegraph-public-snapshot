// tslint:disable

import * as React from "react";
import * as styles from "sourcegraph/components/styles/tabs.css";
import * as classNames from "classnames";

class Tabs extends React.Component<any, any> {
	static propTypes = {
		direction: React.PropTypes.string, // vertical, horizontal
		color: React.PropTypes.string, // blue, purple
		size: React.PropTypes.string, // small, large
		children: React.PropTypes.any,
		className: React.PropTypes.string,
		style: React.PropTypes.object,
	};

	static defaultProps = {
		direction: "horizontal",
		color: "blue",
	};

	_childrenWithProps() {
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

	render(): JSX.Element | null {
		const {direction, className, style} = this.props;
		return <div className={classNames(styles.container, direction === "vertical" ? styles.vertical : styles.horizontal, className)} style={style}>{this._childrenWithProps()}</div>;
	}
}

export default Tabs;
