import * as React from "react";

import { OrganizationCard } from "sourcegraph/components/OrganizationCard";
import { ChevronRight } from "sourcegraph/components/symbols/Primaries";
import { whitespace } from "sourcegraph/components/utils";

const scrollerStyle = {
	overflowY: "scroll",
	maxHeight: 400,
	padding: whitespace[1],
};

interface Props {
	root: GQL.IRoot;
	back: () => void;
	select: (org: string) => () => void;
}

export function OrgSelection(props: Props): JSX.Element {
	if (!props.root || !props.root.currentUser || !props.root.currentUser.githubOrgs) {
		return <div>
			Looks like you don't have any organizations: <a onClick={props.back}>Choose a different plan</a>
		</div>;
	}
	return <div style={{ textAlign: "center" }}>
		<div style={{ padding: whitespace[4] }}>
			<p style={{ maxWidth: 360, margin: "auto", marginBottom: whitespace[3] }}>
				Which organization would you like to use Sourcegraph with?
			</p>
			<div style={scrollerStyle}>
				{props.root.currentUser.githubOrgs.map((org, idx) =>
					<div onClick={props.select(org.name)} key={idx} style={{ marginBottom: 16, textAlign: "left" }}>
						<OrganizationCard name={org.name} desc={org.description} icon={org.avatarURL} />
					</div>
				)}
			</div>
		</div>
		<hr style={{ margin: 0 }} />
		<a onClick={props.back} style={{
			display: "inline-block",
			fontWeight: "bold",
			paddingBottom: whitespace[3],
			paddingTop: whitespace[3],
		}}>Choose a different plan<ChevronRight /></a>
	</div>;
}
