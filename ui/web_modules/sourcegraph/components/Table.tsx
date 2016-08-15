// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/components/styles/table.css";
import * as classNames from "classnames";

interface Props {
	className?: string;
	children?: any;
	bordered?: boolean;
	style?: any;
}

type State = any;

export class Table extends React.Component<Props, State> {
	render(): JSX.Element | null {
		const {className, children, bordered, style} = this.props;

		return (
			<table className={classNames(className, bordered && styles.bordered)} style={style} cellSpacing="0">
				{children}
			</table>
		);
	}
}
