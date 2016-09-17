// tslint:disable: typedef ordered-imports

import * as React from "react";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";

export function withLineColBoundToHash(C) {
	return function WithLineColBoundToHash(props) {
		const rop = props.location && props.location.hash && props.location.hash.startsWith("#L") ?
			RangeOrPosition.parse(props.location.hash.replace(/^#L/, "")) : null;
		// TODO(sqs): make React props use zero-indexed line/col for purity
		const r = rop ? rop.oneIndexed() : null;
		let startLine = r && r.startLine ? r.startLine : null;
		let startCol = r && r.startCol ? r.startCol : null;
		let endLine = r && r.endLine ? r.endLine : null;
		let endCol = r && r.endCol ? r.endCol : null;
		return <C {...props} startLine={startLine} startCol={startCol} endLine={endLine} endCol={endCol} />;
	};
}
