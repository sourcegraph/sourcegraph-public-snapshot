import {hover, style as gStyle} from "glamor";
import * as React from "react";
import {ChevronDown} from "sourcegraph/components/symbols/Zondicons";
import {colors} from "sourcegraph/components/utils";

interface Props {
	color?: string;
	onClick?: any;
	style?: React.CSSProperties;
}

export function ContextMenuIcon ({
	color = colors.blue(),
	onClick,
	style,
}: Props): JSX.Element {

	return <div onClick={onClick}
		{...hover({transform: "scale(1.15)", color: "white"})}
		{...gStyle(Object.assign({},
			{
				backgroundColor: color,
				borderRadius: 3,
				color: colors.black(0.8),
				cursor: "pointer",
				display: "inline-block",
				height: 15,
				width: 15,
				textAlign: "center",
				transition: "all 0.2s cubic-bezier(1, 0, 0, 1)",
			},
			style,
		))}>
		<ChevronDown width={9} style={{
			lineHeight: 0,
			top: -5,
		}} />
	</div>;
};
