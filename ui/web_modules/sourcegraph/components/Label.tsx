// tslint:disable: typedef ordered-imports

import * as React from "react";

import * as styles from "./styles/label.css";

interface Props {
	className?: string;
	style?: any;
	color?: string;
	children?: any;
}

export class Label extends React.Component<Props, any> {
	render(): JSX.Element | null {
		return (
			<span className={this.props.className} style={this.props.style}>
				<span className={colorClasses[this.props.color || ""] || styles.normal}>
					{this.props.children}
				</span>
			</span>
		);
	}
}

const colorClasses = {
	"normal": styles.normal,
	"primary": styles.primary,
	"success": styles.success,
	"info": styles.info,
	"warning": styles.warning,
	"danger": styles.danger,
	"purple": styles.purple,
};
