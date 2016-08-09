// tslint:disable

import * as React from "react";

type Props = {
	className?: string,
	children?: any,
	active?: number,
	styles?: any,
};

class TabPanels extends React.Component<Props, any> {
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
