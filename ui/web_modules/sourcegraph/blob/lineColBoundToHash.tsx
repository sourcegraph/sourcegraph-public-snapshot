// tslint:disable: typedef ordered-imports

import {parseLineRange} from "sourcegraph/blob/lineCol";
import {Helper} from "sourcegraph/blob/BlobLoader";

export const lineColBoundToHash = ({
	reconcileState(state: any, props: any) {
		// pos will contain {start,end}{Line,Col} properties, if any.
		let pos = props.location && props.location.hash && props.location.hash.startsWith("#L") ?
				parseLineRange(props.location.hash.replace(/^#L/, "")) : null;
		state.startLine = pos && pos.startLine ? pos.startLine : null;
		state.startCol = pos && pos.startCol ? pos.startCol : null;
		state.endLine = pos && pos.endLine ? pos.endLine : null;
		state.endCol = pos && pos.endCol ? pos.endCol : null;
	},
} as Helper);
