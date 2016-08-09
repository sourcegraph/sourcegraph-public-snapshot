// tslint:disable

import * as React from "react";

type Props = {
	className?: string,
	children?: any,
	active?: boolean,
	tabPanel?: boolean,
};

class TabPanel extends React.Component<Props, any> {
	static defaultProps = {
		tabPanel: true,
	};

	render(): JSX.Element | null {
		const {className, children, active} = this.props;
		return (
			<div className={className} style={{display: active ? "block" : "none"}}>
				{children}
			</div>
		);
	}
}

export default TabPanel;
