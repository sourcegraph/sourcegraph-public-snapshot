import { hover as gHover } from "glamor";
import * as React from "react";
import { Avatar, Heading, Panel } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props {
	icon?: string;
	name: string;
	desc: string;
	style?: React.CSSProperties;
}

export function OrganizationCard({ desc, icon, name, style }: Props): JSX.Element {

	const panelHover = gHover({ border: `1px ${colors.blueGrayL2()}  solid !important` }).toString();
	const panelSx = {
		border: "1px solid transparent",
		padding: whitespace[3],
		cursor: "pointer",
		...style
	};
	const secondaryText = {
		color: colors.blueGray(),
		display: "inline-block",
		overflow: "hidden",
		textOverflow: "ellipsis",
		whiteSpace: "nowrap",
		width: "calc(100% - 100px)",
		...typography.small,
	};
	const avatarSx = { borderRadius: 3, float: "left", marginRight: whitespace[3] };

	return <Panel hoverLevel="low" className={panelHover} style={panelSx}>
		<Avatar img={icon} style={avatarSx} size="medium" />
		<Heading level={6} compact={true}>{name}</Heading>
		<span style={secondaryText}>{desc}</span>
	</Panel >;
};
