import {hover} from "glamor";
import * as React from "react";
import {Org} from "sourcegraph/api";
import {FlexContainer, OrgLogo} from "sourcegraph/components";
import {GitHubLogo} from "sourcegraph/components/symbols";
import {colors, whitespace} from "sourcegraph/components/utils";
import {Location} from "sourcegraph/Location";

interface Props {
	org: Org;
	location?: Location;
}

const AVATAR_PLACEHOLDER_URL = "https://avatars2.githubusercontent.com/u/10788623?v=3&s=400";

export function OrgCard({org}: Props): JSX.Element {
	return <FlexContainer items="center">
		<div style={{position: "relative"}}>
			<div
			{...hover({fill: colors.white()})}
			style={{
				backgroundColor: colors.coolGray2(),
				borderRadius: "50%",
				position: "absolute",
				padding: 3,
				right: 8,
				top: 2,
				lineHeight: 0,
			}}>
				<GitHubLogo color={colors.white()} width={14} />
			</div>
			<OrgLogo style={{marginRight: whitespace[3]}} size="tiny" img={org.AvatarURL ? org.AvatarURL : AVATAR_PLACEHOLDER_URL} />
		</div>
		{org.Login}
	</FlexContainer>;
}
