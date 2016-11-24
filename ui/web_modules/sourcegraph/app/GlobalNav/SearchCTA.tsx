import { $, merge } from "glamor";
import * as React from "react";
import { Search as SearchIcon } from "sourcegraph/components/symbols";
import { colors } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/index";

export function SearchCTA(props: { style?: any, width: number, content?: string }): JSX.Element {

	const sx = merge(
		{
			display: "inline-block",
			color: colors.coolGray3(),
			padding: whitespace[2],
			marginRight: whitespace[2],
		},
		$(":hover", { color: colors.blueText() }),
		$(":hover svg", { fill: colors.blueText() }),
		props.style ? props.style : {}
	);

	return (
		<div id="SearchCTA-e2e-test" {...sx}>
			<SearchIcon color={colors.coolGray3()} width={props.width} />
			<div style={{ display: "inline", marginLeft: whitespace[2] }}>
				{props.content ? props.content : "Search"}
			</div>
		</div>
	);
};
