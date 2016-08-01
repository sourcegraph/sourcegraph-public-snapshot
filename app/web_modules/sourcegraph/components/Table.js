import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/table.css";

class Table extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		tableStyle: React.PropTypes.string, //  default, bordered
		style: React.PropTypes.object,
	};

	static defaultProps = {
		tableStyle: "default",
	};

	render() {
		const {className, children, tableStyle, style} = this.props;

		return (
			<table className={className} styleName={tableStyle} style={style} cellSpacing="0">
				{children}
			</table>
		);
	}
}

export default CSSModules(Table, styles);
