import * as classNames from "classnames";
import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";

interface Props {
	style?: Object;
	className?: string;
}

export class Toast extends React.Component<Props, {}> {
	static defaultProps: Props = {
		className: classNames(base.pv1, base.ph2, base.ba, base.br2, colors.b__cool_pale_gray, colors.bg_near_white),
	};

	constructor(props: Props) {
		super(props);
		this.state = { alertVisible: true };
	}

	render(): JSX.Element {
		let {children, style, className} = this.props;

		return (
			<div className={className} style={style}>
				{children}
			</div>
		);
	}
}
