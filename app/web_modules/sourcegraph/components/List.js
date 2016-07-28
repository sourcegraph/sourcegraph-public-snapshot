// @flow

import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/list.css";

class List extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		style: React.PropTypes.object,
		listStyle: React.PropTypes.oneOf(["node", "normal"]),
	};

	static defaultProps = {
		listStyle: "normal",
	};

	render() {
		const {className, children, listStyle} = this.props;

		return (
			<ul className={className} styleName={`${listStyle}`} style={this.props.style}>
				{children}
			</ul>
		);
	}
}

export default CSSModules(List, styles, {allowMultiple: true});
