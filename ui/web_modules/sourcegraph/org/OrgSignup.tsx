import * as React from "react";

import { OrganizationCard } from "sourcegraph/components/OrganizationCard";
import { ChevronRight } from "sourcegraph/components/symbols/Primaries";

const scrollerStyle = {
	overflowY: "scroll",
	maxHeight: 400,
	padding: 4,
	marginBottom: 16,
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
	return <div style={{ textAlign: "center", padding: "2rem", paddingBottom: "1rem" }}>
		<div style={{ margin: "20px 55px" }}>
			Which organization would you like to use Sourcegraph with?
		</div>
		<div style={scrollerStyle}>
			{props.root.currentUser.githubOrgs.map((org, idx) =>
				<div onClick={props.select(org.name)} key={idx} style={{ marginBottom: 16 }}>
					<OrganizationCard name={org.name} desc={org.description} icon={org.avatarURL} />
				</div>
			)}
		</div>
		<hr style={{ width: "200%", marginBottom: 16 }} />
		<a onClick={props.back} style={{ fontWeight: "bold" }}>Choose a different plan<ChevronRight /></a>
	</div>;
}
