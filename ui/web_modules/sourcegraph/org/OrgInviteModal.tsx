import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";
import {OrgMember} from "sourcegraph/api";
import {Component} from "sourcegraph/Component";
import {Button, Heading, Table, User} from "sourcegraph/components";
import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import {colors} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/whitespace";

interface Props {
	location: Location;
	members: OrgMember[];
	onInvite: ([]: Array<Object>) => void;
}

interface State {
	selectedInvites: any[];
}

const sx = {
	maxWidth: "800px",
	marginLeft: "auto",
	marginRight: "auto",
};
const rowBorderSx = {
	borderBottomWidth: 1,
	borderColor: colors.coolGray4(0.5),
	borderBottomStyle: "solid",
};

export class OrgInviteModal extends Component<Props, State>  {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

	reconcileState(state: State, props: Props): void {
		state.selectedInvites = this.props.members.filter(
			(member: OrgMember): boolean => {
				return member.Email !== undefined;
			}
		);
	}

	onSubmit(): void {
		let invites: Object[] = [];
		for (let i = 0; i < this.props.members.length; i++) {
			if (this.refs[`checkbox-${i}`]["checked"]) {
				let member = this.props.members[i];
				let rowData = {
					member: member,
					email: this.refs[`email-${i}`]["value"],
				};
				invites.push(rowData);
			}
		}

		// Only trigger dismissal and invite send if there are any to send
		if (invites.length > 0) {
			this.props.onInvite(invites);
		}
	}

	render(): JSX.Element {
		let {members} = this.props;
		return (
			<div>
				<LocationStateModal router={this.context.router} modalName="orgInvite" location={this.props.location}>
					<div style={{paddingTop: "50px"}}>
						<div className={styles.modal} style={sx}>
							<Heading underline="blue" level={3}>Invite Teammates</Heading>
							<p>These are your teammates who are not using Sourcegraph yet.</p>
							<div style={{marginTop: whitespace[3], marginBottom: whitespace[3]}}>
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
											{members && members.map((member, i) =>
												<tr key={i}>
													<td style={rowBorderSx} width="30%">
														<User avatar={member.AvatarURL} email={member.Email} nickname={member.Login} />
													</td>
													<td style={Object.assign({}, rowBorderSx, {textAlign: "left"})} width="60%">
														<input ref={`email-${i}`} style={{boxSizing: "border-box", width: "100%"}} defaultValue={member.Email || ""}/>
													</td>
													<td style={Object.assign({}, rowBorderSx, {textAlign: "center"})} width="10%">
														<input ref={`checkbox-${i}`} type="checkbox" defaultChecked={Boolean(member.Email)}/>
													</td>
												</tr>
											)}
										</tbody>
									</Table>
								<Button onClick={this.onSubmit.bind(this)} style={{float: "right", marginTop: "10px"}} type="submit" color="blue">Invite</Button>
							</div>
						</div>
					</div>
				</LocationStateModal>
			</div>
		);
	}
}
