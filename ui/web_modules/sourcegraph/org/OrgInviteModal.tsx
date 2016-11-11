import * as React from "react";
import {InjectedRouter} from "react-router";
import {Org, OrgMember} from "sourcegraph/api";
import {Button, Heading, Table, User} from "sourcegraph/components";
import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import {colors} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/whitespace";
import {Location} from "sourcegraph/Location";

interface Props {
	location: Location;
	org: Org;
	member: OrgMember | null;
	onInvite: ([]: Array<Object>) => void;
}

interface State {
	isValidForm: boolean;

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

// General Email Regex RFC 5322 Official Standard.
const emailRegex = /^[-a-z0-9~!$%^&*_=+}{\'?]+(\.[-a-z0-9~!$%^&*_=+}{\'?]+)*@([a-z0-9_][-a-z0-9_]*(\.[-a-z0-9_]+)*\.(aero|arpa|biz|com|coop|edu|gov|info|int|mil|museum|name|net|org|pro|travel|mobi|[a-z][a-z])|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}))(:[0-9]{1,5})?$/i;

export class OrgInviteModal extends React.Component<Props, State>  {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

	constructor() {
		super();
		this._validateEmail.bind(this);
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
		let email = this.refs["email"]["value"];
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
		let isValid = emailRegex.test(event.target.value);
		this.setState({
			isValidForm: isValid,
		});
	}

	render(): JSX.Element | null {
		let {member, org} = this.props;
		if (member === null) {
			return null;
		}

		return (
			<div>
				<LocationStateModal router={this.context.router} modalName="orgInvite" location={this.props.location}>
					<div style={{paddingTop: "50px"}}>
						<div className={styles.modal} style={sx}>
							<Heading underline="blue" level={3}>Invite Teammate</Heading>
							<p>Enter a valid email address to invite your teamate to join {org.Login} on Sourcegraph</p>
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
											<td style={Object.assign({}, rowBorderSx, {textAlign: "left"})} width="50%">
												<input onChange={this._validateEmail.bind(this)} type="email" required={true} placeholder="Email address" style={{boxSizing: "border-box", width: "100%"}} defaultValue={member.Email || ""}/>
											</td>
											<td style={rowBorderSx} width="20%">
												<Button onClick={this.onSubmit.bind(this)} disabled={!this.state.isValidForm} style={{float: "right"}} color="blue">Invite</Button>
											</td>
										</tr>
									</tbody>
								</Table>
							</div>
						</div>
					</div>
				</LocationStateModal>
			</div>
		);
	}
}
