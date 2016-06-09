// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/tabs.css";

class Tabs extends React.Component {
	static propTypes = {
		direction: React.PropTypes.string, // vertical, horizontal
		color: React.PropTypes.string, // blue, purple
		size: React.PropTypes.string, // small, large
		children: React.PropTypes.any,
		className: React.PropTypes.string,
	};

	static defaultProps = {
		direction: "horizontal",
		color: "blue",
	};

	_childrenWithProps() {
		return React.Children.map(this.props.children, child =>
			React.cloneElement(child, {
				direction: this.props.direction,
				color: this.props.color,
				size: this.props.size,
			})
		);
	}

	render() {
		const {direction, className} = this.props;
		return <div styleName={`container ${direction}`} className={className}>{this._childrenWithProps()}</div>;
	}
}

export default CSSModules(Tabs, styles, {allowMultiple: true});
