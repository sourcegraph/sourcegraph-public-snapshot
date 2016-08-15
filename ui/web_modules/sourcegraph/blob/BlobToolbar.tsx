// tslint:disable: typedef ordered-imports

import * as React from "react";

import {Component} from "sourcegraph/Component";
import * as s from "sourcegraph/blob/styles/Blob.css";

interface Props {
	repo: string;
	rev?: string;
	commitID?: string;
	path?: string;
}

type State = any;

export class BlobToolbar extends Component<Props, State> {
	reconcileState(state: State, props: Props): void {
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
