// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/components/styles/list.css";
import * as classNames from "classnames";

interface Props {
	className?: string;
	children?: any;
	style?: any;
	listStyle?: string; // node, normal
}

type State = any;

export class List extends React.Component<Props, State> {
	static defaultProps = {
		listStyle: "normal",
	};

	render(): JSX.Element | null {
		const {className, children, listStyle} = this.props;

		return (
			<ul className={classNames(listStyleClasses[listStyle || "normal"] || styles.normal, className)} style={this.props.style}>
				{children}
			</ul>
		);
	}
}

const listStyleClasses = {
	"normal": styles.normal,
	"node": styles.node,
};
