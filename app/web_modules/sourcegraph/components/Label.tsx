// tslint:disable

import * as React from "react";

import CSSModules from "react-css-modules";
import * as styles from "./styles/label.css";

class Label extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		style: React.PropTypes.object,
		color: React.PropTypes.string,
		outline: React.PropTypes.bool,
		children: React.PropTypes.any,
	};

	render(): JSX.Element | null {
		return (
			<span className={this.props.className} style={this.props.style}>
				<span styleName={`${this.props.outline ? "outline_" : ""}${this.props.color || "normal"}`}>
					{this.props.children}
				</span>
			</span>
		);
	}
}

export default CSSModules(Label, styles, {allowMultiple: true});
