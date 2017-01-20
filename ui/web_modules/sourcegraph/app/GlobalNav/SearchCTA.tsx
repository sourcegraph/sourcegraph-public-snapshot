import { $, merge } from "glamor";
import * as React from "react";
import { Search } from "sourcegraph/components/symbols/Primaries";
import { colors } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/index";

export function SearchCTA(props: { style?: any, width: number, content?: string }): JSX.Element {

	const sx = merge(
		{
			display: "inline-block",
			color: colors.blueGray(),
			padding: whitespace[2],
			marginRight: whitespace[2],
		},
		$(":hover", { color: colors.blue() }),
		$(":hover svg", { fill: colors.blue() }),
		props.style ? props.style : {}
	);

	return (
		<div id="SearchCTA-e2e-test" {...sx}>
			<Search color={colors.blueGray()} width={props.width} />
			<div style={{ display: "inline", marginLeft: whitespace[1] }}>
				{props.content ? props.content : "Search"}
			</div>
		</div>
	);
};
