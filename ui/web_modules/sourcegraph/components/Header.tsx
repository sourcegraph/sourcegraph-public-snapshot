// tslint:disable

import * as React from "react";
import * as styles from "./styles/header.css";
import Loader from "./Loader";

type Props = {
	title: string,
	subtitle?: string,
	loading?: boolean,
};

class Header extends React.Component<Props, any> {
	render(): JSX.Element | null {
		return (
			<div className={styles.container}>
				<div className={styles.cloning_title}>{this.props.title}</div>
				<div className={styles.cloning_subtext}>{this.props.loading ? <Loader /> : this.props.subtitle}</div>
			</div>
		);
	}
}

export default Header;
