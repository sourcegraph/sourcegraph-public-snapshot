import { hover as gHover } from "glamor";
import * as React from "react";
import { Avatar, Heading, Panel } from "sourcegraph/components";
import { UserWomanAlternate } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props {
	icon?: string;
	name: string;
	userCount: number;
	hover?: boolean;
	style?: React.CSSProperties;
}

export function OrganizationCard({ hover, icon, name, style, userCount }: Props): JSX.Element {

	const panelHover = gHover({ border: `1px ${colors.blueGrayL2()}  solid !important` }).toString();
	const userCountSx = {
		...{ color: colors.blueGray() },
		...typography.small,
	};
	const panelSx = {
		...{ border: "1px solid transparent", padding: whitespace[3] },
		...style
	};
	const userIconSx = { marginRight: whitespace[1], verticalAlign: "text-bottom" };
	const avatarSx = { float: "left", marginRight: whitespace[3] };

	return <Panel hoverLevel="low" className={hover ? panelHover : ""} style={panelSx}>
		<Avatar img={icon} style={avatarSx} size="medium" />
		<Heading level={6} compact={true}>{name}</Heading>
		<span style={userCountSx}>
			<UserWomanAlternate color={colors.blueGrayL1()} width={16} style={userIconSx} />
			{userCount} users
		</span>
	</Panel >;
};
