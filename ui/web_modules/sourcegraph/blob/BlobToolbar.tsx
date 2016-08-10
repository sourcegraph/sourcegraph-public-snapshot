// tslint:disable: typedef ordered-imports curly

import * as React from "react";

import {Component} from "sourcegraph/Component";
import * as s from "sourcegraph/blob/styles/Blob.css";

export class BlobToolbar extends Component<Props, any> {
	reconcileState(state, props) {
		state.repo = props.repo;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.path = props.path || null;
	}

	render(): JSX.Element | null {
		return (
			<div className={s.toolbar}>
				<div className="actions">
				</div>
			</div>
		);
	}
}

type Props = {
	repo: string,
	rev?: string,
	commitID?: string,
	path?: string,
};
