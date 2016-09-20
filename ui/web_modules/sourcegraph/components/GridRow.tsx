import * as React from "react";

import {Base} from "sourcegraph/components/Base";
import {clearFix} from "sourcegraph/components/utils/layout";

interface Props {
	style: Object;
	children: Array<JSX.Element>;
}

export function GridRow(props: Props): JSX.Element {
	return <Base
		{...props}
		{...clearFix}
		style={props.style}>
			{props.children}
		</Base>;
}
