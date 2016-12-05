import * as classNames from "classnames";
import * as React from "react";

import * as styles from "sourcegraph/components/styles/table.css";
interface Props {
	className?: string;
	children?: any;
	bordered?: boolean;
	style?: any;
}

export class Table extends React.Component<Props, {}> {
	render(): JSX.Element | null {
		const {className, children, bordered, style} = this.props;

		return (
			<table className={classNames(className, bordered ? styles.bordered : null)} style={style} cellSpacing="0">
				{children}
			</table>
		);
	}
}
