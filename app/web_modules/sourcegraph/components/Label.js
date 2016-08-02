import * as React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/label.css";

class Label extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		style: React.PropTypes.object,
		color: React.PropTypes.string,
		outline: React.PropTypes.bool,
		children: React.PropTypes.any,
	};

	render() {
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
