// tslint:disable: typedef ordered-imports

import * as React from "react";
import { parseLineRange } from "sourcegraph/blob/lineCol";

export function withLineColBoundToHash(C) {
	return function WithLineColBoundToHash(props) {
		// pos will contain {start,end}{Line,Col} properties, if any.
		let pos = props.location && props.location.hash && props.location.hash.startsWith("#L") ?
			parseLineRange(props.location.hash.replace(/^#L/, "")) : null;
		let startLine = pos && pos.startLine ? pos.startLine : null;
		let startCol = pos && pos.startCol ? pos.startCol : null;
		let endLine = pos && pos.endLine ? pos.endLine : null;
		let endCol = pos && pos.endCol ? pos.endCol : null;
		return <C {...props} startLine={startLine} startCol={startCol} endLine={endLine} endCol={endCol} />;
	};
}
