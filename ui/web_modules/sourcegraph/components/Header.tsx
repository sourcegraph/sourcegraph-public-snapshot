// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "./styles/header.css";
import {Loader} from "./Loader";

interface Props {
	title: string;
	subtitle?: string;
	loading?: boolean;
}

type State = any;

export class Header extends React.Component<Props, State> {
	render(): JSX.Element | null {
		return (
			<div className={styles.container}>
				<div className={styles.cloning_title}>{this.props.title}</div>
				<div className={styles.cloning_subtext}>{this.props.loading ? <Loader /> : this.props.subtitle}</div>
			</div>
		);
	}
}
