import * as React from "react";
import {OrgMember} from "sourcegraph/api";
import {Button, Heading, Table, User} from "sourcegraph/components";
import {colors} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/whitespace";

interface Props {
	members: OrgMember[];
	inviteClicked: (member: OrgMember) => void;
	sentInvites: Array<String>;
}

export class OrgMembersTable extends React.Component<Props, {}> {
	_inviteSelected(member: OrgMember): void {
		if (this.props.sentInvites.indexOf(member.Login) === -1) {
			this.props.inviteClicked(member);
		}
	}

	render(): JSX.Element {
		let {members} = this.props;

		if (members.length === 0) {
			return <div style={{marginTop: whitespace[3], marginBottom: whitespace[3]}}>
				<p>Looks like your organization is empty. Invite some of your users to join!</p>
			</div>;
		}

		const rowBorderSx = {
			borderBottomWidth: 1,
			borderColor: colors.coolGray4(0.5),
			borderBottomStyle: "solid",
		};

		return <div style={{marginTop: whitespace[3], marginBottom: whitespace[3]}}>
			<Table style={{width: "100%"}}>
				<thead>
					<tr>
						<td style={rowBorderSx}>
							<Heading level={6}>
								Organization member
							</Heading>
						</td>
						<td
							style={Object.assign({},
								rowBorderSx,
								{
									textAlign: "center",
									padding: "12px 0",
									whiteSpace: "nowrap",
								})
							}>
						</td>
					</tr>
				</thead>
				<tbody>
					{members.map((member, i) =>
						<tr key={i}>
							<td style={rowBorderSx}>
								<User avatar={member.AvatarURL} email={member.Email} nickname={member.Login} />
							</td>
							<td style={Object.assign({}, rowBorderSx, {textAlign: "center"})} width="20%">
								{!member.SourcegraphUser && (member.Invite || (this.props.sentInvites.indexOf(member.Login) > -1) ? <div style={{fontStyle: "italic"}}>Invite sent</div> : <Button color="blue" disabled={!(member.CanInvite || !member.Invite)} onClick={(e) => {this._inviteSelected(member);}}>Invite</Button>)}
								{member.SourcegraphUser && <div style={{fontStyle: "italic"}}>Member</div>}
							</td>
						</tr>
					)}
				</tbody>
			</Table>
		</div>;
	}
};
