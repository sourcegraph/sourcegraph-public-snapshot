// tslint:disable

import * as React from "react";

class TabPanels extends React.Component<any, any> {
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
		return React.Children.map(this.props.children, (child: React.ReactElement<any>, i) => {
			if (child.props.tabPanel) {
				if (this.props.active === i) return React.cloneElement(child, {active: true});
				return React.cloneElement(child, {active: false});
			}
			return React.cloneElement(child);
		});
	}

	render(): JSX.Element | null {
		const {className, styles} = this.props;
		return (
			<div className={className} style={styles}>
				{this._childrenWithProps()}
			</div>
		);
	}
}

export default TabPanels;
