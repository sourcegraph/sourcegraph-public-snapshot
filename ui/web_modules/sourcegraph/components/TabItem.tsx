// tslint:disable

import * as React from "react";
import {Icon} from "./Icon";
import * as styles from "sourcegraph/components/styles/tabItem.css";
import * as classNames from "classnames";

type Props = {
	className?: string,
	children?: any,
	hideMobile?: boolean,
	active?: boolean,
	color?: string, // blue, purple
	size?: string, // small, large
	icon?: string | JSX.Element,
	direction?: string,
	tabItem?: boolean,
};

export class TabItem extends React.Component<Props, any> {
	static defaultProps = {
		active: false,
		color: "blue",
		direction: "horizontal",
		tabItem: true,
	};

	render(): JSX.Element | null {
		const {size, children, hideMobile, active, color, icon, direction} = this.props;
		return (
			<span
				className={classNames(sizeClasses[size || "small"], hideMobile && styles.hidden_s, active ? styles.active : styles.inactive, colorClasses[color || "blue"] || styles.blue, direction === "vertical" ? styles.vertical : styles.horizontal)}>
				{icon && typeof icon === "string" && <Icon icon={`${icon}-blue`} height="14px" width="auto" className={classNames(styles.icon, !active && styles.hide)}/>}
				{icon && typeof icon === "string" && <Icon icon={`${icon}-gray`} height="14px" width="auto" className={classNames(styles.icon, active && styles.hide)}/>}
				{icon && typeof icon !== "string" && React.cloneElement(icon, {className: active ? `${styles.component_icon} ${styles.active} ${colorClasses[color || "blue"]}` : `${styles.component_icon} ${styles.inactive}`})}
				{children}
			</span>
		);
	}
}

const sizeClasses = {
	"small": styles.small,
	"large": styles.large,
};

const colorClasses = {
	"blue": styles.blue,
	"purple": styles.purple,
};
