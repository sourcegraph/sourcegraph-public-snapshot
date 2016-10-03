import {$, merge} from "glamor";
import * as React from "react";
import {Base} from "sourcegraph/components";
import {Search as SearchIcon} from "sourcegraph/components/symbols";
import {colors} from "sourcegraph/components/utils";

export function SearchCTA(props: {style?: any, width: number, content?: string}): JSX.Element {

	const sx = merge(
		{
			display: "inline-block",
			color: colors.coolGray3(),
		},
		$(":hover", { color: colors.blueText() }),
		$(":hover svg", { fill: colors.blueText() }),
		props.style ? props.style : {}
	);

	return(
		<Base id="SearchCTA-e2e-test" p={2} mr={2} {...sx}>
			<SearchIcon color={colors.coolGray3()} width={props.width} />
			<Base ml={2} style={{display: "inline"}}>
				{props.content ? props.content : "Search"}
			</Base>
		</Base>
	);
};
