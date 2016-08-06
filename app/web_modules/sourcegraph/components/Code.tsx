// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/code_2.css";

class Code extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		style: React.PropTypes.object,
	};

	render(): JSX.Element | null {
		const {className, children, style} = this.props;
		return <span className={className} style={style} styleName="code">{children}</span>;
	}
}

export default CSSModules(Code, styles);
