// tslint:disable

import * as React from "react";
import * as styles from "./styles/code_2.css";
import * as classNames from "classnames";

type Props = {
	className?: string,
	children?: any,
	style?: any,
};

class Code extends React.Component<Props, any> {
	render(): JSX.Element | null {
		const {className, children, style} = this.props;
		return <span className={classNames(className, styles.code)} style={style}>{children}</span>;
	}
}

export default Code;
