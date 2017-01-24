import { css, insertGlobal } from "glamor";
import * as React from "react";
import { User } from "sourcegraph/components";
import { Document } from "sourcegraph/components/symbols/Primaries";
import { colors, paddingMargin } from "sourcegraph/components/utils";

const rowSx = paddingMargin.margin("y", 1);

const sx = css(
	paddingMargin.padding("x", 3),
	paddingMargin.padding("y", 2),
	{
		borderTop: "1px solid transparent",
		borderBottom: `1px solid ${colors.blueGrayL1(0.2)}`,
		color: colors.text(),
		cursor: "pointer",
		marginTop: -1,
	},
	{
		":hover": {
			backgroundColor: colors.blueL3(),
			borderColor: colors.blueL2(0.7),
		}
	},
	{
		":active": {
			backgroundColor: colors.blue(),
			color: "white",
		}
	},
);

// Selected needs to look at the parent .monaco-tree-row
insertGlobal(`#reference-tree .monaco-tree-row.selected [data-${sx}]`, {
	backgroundColor: colors.blue(),
	borderColor: "white",
	boxShadow: `inset 0 1px 3px ${colors.black(0.25)} `,
	color: "white",
});
insertGlobal(`#reference-tree .monaco-tree-row.selected [data-${sx}]:hover`, {
	borderColor: "white",
});
insertGlobal(`#reference-tree .monaco-tree-row.selected [data-${sx}] div`, {
	color: "white",
});

export function ReferenceCard({ fnSignature, authorName, avatar, date, fileName, line }: {
	fnSignature: string;
	authorName?: string;
	avatar?: string;
	date?: string;
	fileName: string;
	line: number;
}): JSX.Element {
	return <div {...sx}>
		<code style={Object.assign({
			WebkitFontSmoothing: "auto",
			fontSize: 14,
			display: "block",
			width: "100%",
			textOverflow: "ellipsis",
			overflow: "hidden",
			wordWrap: "nowrap",
		}, rowSx)} dangerouslySetInnerHTML={{ __html: fnSignature }}></code>
		<div {...css({ color: colors.blueGrayD1() }) }>
			{authorName && <div style={rowSx}>
				<User simple={true} nickname={`${authorName} · ${date}`} avatar={avatar} style={{ display: "inline-block" }} />
			</div>}
			<div style={rowSx}>
				<Document style={{ opacity: 0.75, marginRight: 6 }} width={19} />
				{fileName}: Line {line}
			</div>
		</div>
	</div>;
};
