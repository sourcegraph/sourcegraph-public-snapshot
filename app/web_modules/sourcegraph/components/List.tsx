// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/list.css";

class List extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		style: React.PropTypes.object,
		listStyle: React.PropTypes.oneOf(["node", "normal"]),
	};

	static defaultProps = {
		listStyle: "normal",
	};

	render(): JSX.Element | null {
		const {className, children, listStyle} = this.props;

		return (
			<ul className={className} styleName={`${listStyle}`} style={this.props.style}>
				{children}
			</ul>
		);
	}
}

export default CSSModules(List, styles, {allowMultiple: true});
