import * as React from "react";

import {Loader} from "sourcegraph/components/Loader";
import * as styles from "sourcegraph/components/styles/header.css";

interface Props {
	title: string;
	subtitle?: string;
	loading?: boolean;
}

export class Header extends React.Component<Props, {}> {
	render(): JSX.Element | null {
		return (
			<div className={styles.container}>
				<div className={styles.cloning_title}>{this.props.title}</div>
				<div className={styles.cloning_subtext}>{this.props.loading ? <Loader /> : this.props.subtitle}</div>
			</div>
		);
	}
}
