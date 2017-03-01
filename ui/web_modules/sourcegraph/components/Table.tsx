import * as classNames from "classnames";
import { css } from "glamor";
import * as React from "react";
import { colors, whitespace } from "sourcegraph/components/utils";

interface Props {
	className?: string;
	children?: React.ReactNode[];
	bordered?: boolean;
	style?: React.CSSProperties;
}

const defSx = css({
	"& td": { padding: `${whitespace[3]} ${whitespace[3]} ${whitespace[3]} 0` },
	"& tbody td": { borderBottom: `1px solid ${colors.blueGrayL2()}` },
	"& thead td": {
		fontWeight: "bold",
		borderBottom: `2px solid ${colors.blueGrayL2()}`,
	}
}).toString();

const borderedSx = css({
	border: `1px solid ${colors.blueGrayL2()}`,
	borderBottom: 0,
	"& td": { padding: whitespace[3] },
}).toString();

export function Table({ className, children, bordered, style }: Props): JSX.Element {
	return (
		<table
			className={classNames(defSx, bordered ? borderedSx : "", className)}
			style={style}
			cellSpacing="0">
			{children}
		</table>
	);
}
