// tslint:disable

import * as React from "react";
import Icon from "./Icon";
import * as styles from "sourcegraph/components/styles/tabItem.css";
import * as classNames from "classnames";

class TabItem extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		hideMobile: React.PropTypes.bool,
		active: React.PropTypes.bool,
		color: React.PropTypes.string, // blue, purple
		size: React.PropTypes.string, // small, large
		icon: React.PropTypes.oneOfType([React.PropTypes.string, React.PropTypes.element]),
		direction: React.PropTypes.string,
		tabItem: React.PropTypes.bool,
	};

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
				className={classNames(sizeClasses[size], hideMobile && styles.hidden_s, active ? styles.active : styles.inactive, colorClasses[color] || styles.blue, direction === "vertical" ? styles.vertical : styles.horizontal)}>
				{icon && typeof icon === "string" && <Icon icon={`${icon}-blue`} height="14px" width="auto" className={classNames(styles.icon, !active && styles.hide)}/>}
				{icon && typeof icon === "string" && <Icon icon={`${icon}-gray`} height="14px" width="auto" className={classNames(styles.icon, active && styles.hide)}/>}
				{icon && typeof icon !== "string" && React.cloneElement(icon, {className: active ? `${styles.component_icon} ${styles.active} ${colorClasses[color]}` : `${styles.component_icon} ${styles.inactive}`})}
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

export default TabItem;
