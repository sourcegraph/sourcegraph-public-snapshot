import * as React from "react";

import { Org, OrgMember } from "sourcegraph/api";
import { Router } from "sourcegraph/app/router";
import { Button, Heading, Table, User } from "sourcegraph/components";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { colors } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/whitespace";

interface Props {
	org: Org;
	member: OrgMember | null;
	onInvite: ([]: Array<Object>) => void;
}

interface State { isValidForm: boolean; }

const rowBorderSx = {
	borderBottomWidth: 1,
	borderColor: colors.blueGrayL3(0.5),
	borderBottomStyle: "solid",
};

// General Email Regex RFC 5322 Official Standard.
const emailRegex = /^[-a-z0-9~!$%^&*_=+}{\'?]+(\.[-a-z0-9~!$%^&*_=+}{\'?]+)*@([a-z0-9_][-a-z0-9_]*(\.[-a-z0-9_]+)*\.(aero|arpa|biz|com|coop|edu|gov|info|int|mil|museum|name|net|org|pro|travel|mobi|[a-z][a-z])|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}))(:[0-9]{1,5})?$/i;

export class OrgInviteModal extends React.Component<Props, State>  {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };
	email: HTMLInputElement;

	constructor() {
		super();
		this._validateEmail.bind(this);
		this.state = { isValidForm: false };
	}

	componentDidMount(): void {
		document.body.addEventListener("keydown", this._shouldSubmitInvite.bind(this));
	}

	componentWillUnmount(): void {
		document.body.removeEventListener("keydown", this._shouldSubmitInvite);
	}

	_shouldSubmitInvite(event: KeyboardEvent & Event): void {
		if (event.keyCode === 13) { // Enter.
			if (this.state.isValidForm) {
				this.onSubmit();
			}
			event.preventDefault();
		}
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
	}

	onSubmit(): void {
		let invites: Object[] = [];
		let email = this.email["value"];
		if (emailRegex.test(email)) {
			let member = this.props.member;
			let rowData = {
				member: member,
				email: email,
			};
			invites.push(rowData);
		}

		// Only trigger dismissal and invite send if there are any to send
		if (invites.length > 0) {
			this.props.onInvite(invites);
			this.setState({
				isValidForm: false,
			});
		}
	}

	_validateEmail(event: React.FormEvent<HTMLInputElement>): void {
		let isValid = emailRegex.test(event.currentTarget.value);
		this.setState({
			isValidForm: isValid,
		});
	}

	render(): JSX.Element | null {
		let { member, org } = this.props;
		if (member === null) {
			return null;
		}

		return <LocationStateModal modalName="orgInvite" title="Invite teammate" style={{ maxWidth: 800 }}>
			<p>Enter a valid email address to invite your teamate to join {org.Login} on Sourcegraph</p>
			<div style={{ marginTop: whitespace[3], marginBottom: whitespace[3] }}>
				<Table style={{ width: "100%" }}>
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
										padding: "15px 0",
										whiteSpace: "nowrap",
									})
								}>
							</td>
							<td
								style={Object.assign({},
									rowBorderSx,
									{
										textAlign: "center",
										padding: "15px 0",
										whiteSpace: "nowrap",
									})
								}>
							</td>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td style={rowBorderSx} width="30%">
								<User avatar={member.AvatarURL} email={member.Email} nickname={member.Login} />
							</td>
							<td style={Object.assign({}, rowBorderSx, { textAlign: "left" })} width="50%">
								<input
									onChange={this._validateEmail.bind(this)}
									type="email"
									required={true}
									placeholder="Email address"
									ref={(el) => this.email = el}
									style={{ boxSizing: "border-box", width: "100%" }}
									defaultValue={member.Email || ""} />
							</td>
							<td style={rowBorderSx} width="20%">
								<Button onClick={this.onSubmit.bind(this)} disabled={!this.state.isValidForm} style={{ float: "right" }} color="blue">Invite</Button>
							</td>
						</tr>
					</tbody>
				</Table>
			</div>
		</LocationStateModal>;
	}
}
