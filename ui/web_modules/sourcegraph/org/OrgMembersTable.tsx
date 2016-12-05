import * as React from "react";
import { OrgMember } from "sourcegraph/api";
import { Button, Heading, Table, User } from "sourcegraph/components";
import { colors, whitespace } from "sourcegraph/components/utils";

interface Props {
	members: OrgMember[];
	inviteClicked: (member: OrgMember) => void;
	sentInvites: Array<String>;
}

export function OrgMembersTable({members, inviteClicked, sentInvites}: Props): JSX.Element {
	function _inviteSelected(member: OrgMember): void {
		if (sentInvites.indexOf(member.Login) === -1) {
			inviteClicked(member);
		}
	}

	if (members.length === 0) {
		return <div style={{ marginTop: whitespace[3], marginBottom: whitespace[3] }}>
			<p>Looks like your organization is empty. Invite some of your users to join!</p>
		</div>;
	}

	const rowSx = {
		borderBottomWidth: 1,
		borderColor: colors.coolGray4(0.5),
		borderBottomStyle: "solid",
		paddingBottom: whitespace[2],
		paddingTop: whitespace[2],
	};

	const memberCellSx = Object.assign({ textAlign: "center" }, rowSx);

	return <div style={{ marginBottom: whitespace[3] }}>
		<Table style={{ width: "100%" }}>
			<thead>
				<tr>
					<td style={rowSx}>
						<Heading level={6}>Organization member</Heading>
					</td>
					<td style={memberCellSx}></td>
				</tr>
			</thead>
			<tbody>
				{members.map((member, i) =>
					<tr key={i}>
						<td style={rowSx}>
							<User avatar={member.AvatarURL} email={member.Email} nickname={member.Login} />
						</td>
						<td style={memberCellSx} width="20%">
							{!member.SourcegraphUser && (member.Invite || (sentInvites.indexOf(member.Login) > -1)
								? "Invite sent"
								: <Button size="small" color="blue" disabled={!(member.CanInvite || !member.Invite)} onClick={(e) => { _inviteSelected(member); } }>Invite</Button>)}
							{member.SourcegraphUser && "Member"}
						</td>
					</tr>
				)}
			</tbody>
		</Table>
	</div>;
};
