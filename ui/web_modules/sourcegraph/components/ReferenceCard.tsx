import { css } from "glamor";
import * as React from "react";
import { User } from "sourcegraph/components";
import { Document } from "sourcegraph/components/symbols/Primaries";
import { colors, paddingMargin } from "sourcegraph/components/utils";

const rowSx = paddingMargin.margin("y", 1);

export function ReferenceCard({ fnSignature, authorName, avatar, date, fileName, line }: {
	fnSignature: string;
	authorName?: string;
	avatar?: string;
	date?: string;
	fileName: string;
	line: number;
}): JSX.Element {
	return <div {...css(
		paddingMargin.padding("x", 3),
		paddingMargin.padding("y", 2),
		{
			borderTop: "1px solid transparent",
			borderBottom: `1px solid ${colors.blueGrayL1(0.2)}`,
			color: colors.blueGray(),
			marginTop: -1,
		},
		{
			":hover": {
				backgroundColor: colors.blueL3(),
				borderColor: colors.blueL2(0.7),
			}
		},
	) }>
		<code style={Object.assign({
			color: colors.text(),
			WebkitFontSmoothing: "auto",
			fontSize: 14,
			display: "block",
		}, rowSx)}>{fnSignature}</code>
		{authorName && <div style={rowSx}>
			<User simple={true} nickname={`${authorName} · ${date}`} avatar={avatar} style={{ display: "inline-block" }} />
		</div>}
		<div style={rowSx}>
			<Document style={{ color: colors.blueGrayL1(), marginRight: 6 }} width={19} />
			{fileName}: Line {line}
		</div>
	</div>;
};
