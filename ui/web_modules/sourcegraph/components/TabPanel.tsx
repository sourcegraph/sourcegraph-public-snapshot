import * as React from "react";

interface Props {
	className?: string;
	children?: any;
	active?: boolean;
	tabPanel?: boolean;
}

export class TabPanel extends React.Component<Props, {}> {
	static defaultProps: Props = {
		tabPanel: true,
	};

	render(): JSX.Element | null {
		const {className, children, active} = this.props;
		return (
			<div className={className} style={{ display: active ? "block" : "none" }}>
				{children}
			</div>
		);
	}
}
