import { hover as gHover } from "glamor";
import * as React from "react";
import { Avatar, Heading, Input, Panel } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props {
	icon?: string;
	name: string;
	desc: string;
	style?: React.CSSProperties;
}

const panelHover = gHover({ border: `1px ${colors.blueGrayL2()}  solid !important` }).toString();
const panelSx = {
	border: "1px solid transparent",
	padding: whitespace[3],
	cursor: "pointer",
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

export function OrganizationCard({ desc, icon, name, style }: Props): JSX.Element {
	return <Panel hoverLevel="low" className={panelHover} style={{ ...panelSx, ...style }}>
		<Avatar img={icon} style={avatarSx} size="medium" />
		<Heading level={6} compact={true}>{name}</Heading>
		<span style={secondaryText}>{desc}</span>
	</Panel >;
};

export function OrgSeatsCard({ seats, org, onChange }: {
	seats: string,
	org: GQL.IOrganization,
	onChange: (ev: React.FormEvent<HTMLInputElement>) => void,
}): JSX.Element {
	const count = parseInt(seats, 10);
	let errorText = "";
	if (count < 1) {
		errorText = "You must have at least one seat.";
	}
	return <Panel hoverLevel="low" style={{ ...panelSx, cursor: "default", marginTop: 32, marginRight: 32, marginLeft: 32 }}>
		<Avatar img={org.avatarURL} style={avatarSx} size="medium" />
		<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
			<div style={{ textAlign: "left" }}>
				<Heading level={6} compact={true}>{org.name}</Heading>
			</div>
			<Input style={{ width: 80 }}
				containerStyle={{ marginBottom: 0 }}
				errorText={errorText}
				min="1" type="number"
				onChange={onChange} value={seats} />
		</div>
	</Panel>;
}
