// @flow

import React from "react";

class TabPanels extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		active: React.PropTypes.number,
		styles: React.PropTypes.object,
	};

	static defaultProps = {
		active: 0,
	};

	_childrenWithProps(): any {
		return React.Children.map(this.props.children, (child, i) => {
			if (child.type.name === "TabPanel") {
				console.log("TabPanel If Statement", child);
				if (this.props.active === i) return React.cloneElement(child, {active: true});
				return React.cloneElement(child, {active: false});
			}
			console.log("TabPanel Outside", child);
			return React.cloneElement(child);
		});
	}

	render() {
		const {className, styles} = this.props;
		return (
			<div className={className} style={styles}>
				{this._childrenWithProps()}
			</div>
		);
	}
}

export default TabPanels;
