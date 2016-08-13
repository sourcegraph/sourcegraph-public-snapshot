// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "./styles/code_2.css";
import * as classNames from "classnames";

interface Props {
	className?: string;
	children?: any;
	style?: any;
}

type State = any;

export class Code extends React.Component<Props, State> {
	render(): JSX.Element | null {
		const {className, children, style} = this.props;
		return <span className={classNames(className, styles.code)} style={style}>{children}</span>;
	}
}
