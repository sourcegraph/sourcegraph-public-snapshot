// tslint:disable

import * as React from "react";
import * as styles from "./styles/list.css";
import * as classNames from "classnames";

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
			<ul className={classNames(listStyleClasses[listStyle] || styles.normal, className)} style={this.props.style}>
				{children}
			</ul>
		);
	}
}

const listStyleClasses = {
	"normal": styles.normal,
	"node": styles.node,
};

export default List;
