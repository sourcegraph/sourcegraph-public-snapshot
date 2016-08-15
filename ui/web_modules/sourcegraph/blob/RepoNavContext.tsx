// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";

import {Component} from "sourcegraph/Component";

import {urlTo} from "sourcegraph/util/urlTo";
import {breadcrumb} from "sourcegraph/util/breadcrumb";

import * as styles from "sourcegraph/components/styles/breadcrumb.css";

interface Props {
	params: any;
}

type State = any;

export class RepoNavContext extends Component<Props, State> {
	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
	}

	render(): JSX.Element | null {
		let blobPath = this.props.params.splat[1];
		if (!blobPath) {
			return null;
		}
		let pathParts = blobPath.split("/");
		let pathBreadcrumb = breadcrumb(
			`/${blobPath}`,
			(i) => <span key={i} className={styles.sep}>/</span>,
			(path, component, i, isLast) => (
				<Link to={isLast ?
					urlTo("blob", Object.assign({}, this.state.params)) :
					urlTo("tree", Object.assign({}, this.state.params, {
						splat: [this.state.params.splat[0], pathParts.slice(0, i).join("/")],
					}))}
					key={i}
					className={isLast ? styles.active : styles.inactive}>
					{component}
				</Link>
			),
		);

		return (
			<span>{pathBreadcrumb}</span>
		);
	}
}
