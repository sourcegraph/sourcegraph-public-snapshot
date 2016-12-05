import * as classNames from "classnames";
import * as React from "react";

import * as styles from "sourcegraph/components/styles/menu.css";

interface Props {
	children?: any;
	className?: string;
	style?: any;
}

export class Menu extends React.Component<Props, {}> {
	renderMenuItems(): JSX.Element[] {
		return React.Children.map(this.props.children, (ch: React.ReactElement<any>) =>
			ch && <div key={ch.props} className={roleStyle(ch.props.role)}>{React.cloneElement(ch)}</div>
		);
	}

	render(): JSX.Element | null {
		return <div className={classNames(this.props.className, styles.container)} style={this.props.style}>{this.renderMenuItems()}</div>;
	}
}

function roleStyle(role: string): string {
	switch (role) {
		case "menu_item":
			return styles.menu_item;
		case "divider":
			return styles.divider;
		case "inactive":
			return styles.inactive;
		default:
			return styles.inactive;
	}
}
