// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/table.css";

class Table extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		tableStyle: React.PropTypes.string, //  default, bordered
		style: React.PropTypes.object,
	};

	static defaultProps = {
		tableStyle: "normal",
	};

	render(): JSX.Element | null {
		const {className, children, tableStyle, style} = this.props;

		return (
			<table className={className} styleName={tableStyle} style={style} cellSpacing="0">
				{children}
			</table>
		);
	}
}

export default CSSModules(Table, styles);
