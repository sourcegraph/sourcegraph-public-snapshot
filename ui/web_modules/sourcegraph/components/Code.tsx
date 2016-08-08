// tslint:disable

import * as React from "react";
import * as styles from "./styles/code_2.css";
import * as classNames from "classnames";

class Code extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		style: React.PropTypes.object,
	};

	render(): JSX.Element | null {
		const {className, children, style} = this.props;
		return <span className={classNames(className, styles.code)} style={style}>{children}</span>;
	}
}

export default Code;
