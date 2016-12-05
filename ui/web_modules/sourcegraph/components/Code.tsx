import * as classNames from "classnames";
import * as React from "react";

import * as styles from "sourcegraph/components/styles/code_2.css";

interface Props {
	className?: string;
	children?: any;
	style?: any;
}

export class Code extends React.Component<Props, {}> {
	render(): JSX.Element | null {
		const {className, children, style} = this.props;
		return <span className={classNames(className, styles.code)} style={style}>{children}</span>;
	}
}
