// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Helper} from "sourcegraph/blob/BlobLoader";
import {BlobMargin} from "sourcegraph/blob/BlobMargin";
import {BlobPos} from "sourcegraph/def/DefActions";

export const withBlobMargin = ({
	reconcileState(state: any, props: any): void {
		let p: BlobPos = {
			repo: props.repo,
			commit: props.commitID,
			file: state.path,
			line: state.startLine - 1,
			character: state.startCol,
		};
		state.startPos = p.repo !== null && p.commit !== null && p.file !== null && p.line !== null && p.character !== null ? p : null;
	},

	renderProps(state) {
		let p: BlobPos | null = state.startPos;
		return p !== null ? {
			children: <BlobMargin pos={p} />,
		} : null;
	},
} as Helper);
