import { $, merge } from "glamor";
import * as React from "react";
import { Keyboard, Search } from "sourcegraph/components/symbols/Primaries";
import { colors } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/index";

export const globalNavCtaStyle = merge(
	{
		display: "inline-block",
		color: colors.blueGray(),
		padding: whitespace[2],
		paddingTop: 10,
		marginRight: whitespace[2],
	},
	$(":hover", { color: colors.blue() }),
	$(":hover svg", { fill: colors.blue() }),
);

export function SearchCTA(props: { width: number }): JSX.Element {
	return (
		<div id="SearchCTA-e2e-test" {...globalNavCtaStyle}>
			<Search color={colors.blueGray()} width={props.width} />
			<div style={{ display: "inline", marginLeft: whitespace[1] }}>
				Search
			</div>
		</div>
	);
};

export function ShortcutCTA(props: { width: number }): JSX.Element {
	return (
		<div {...globalNavCtaStyle}>
			<Keyboard color={colors.blueGray()} width={props.width} />
			<div style={{ display: "inline", marginLeft: whitespace[1] }}>
				Shortcuts
			</div>
		</div>
	);
};
