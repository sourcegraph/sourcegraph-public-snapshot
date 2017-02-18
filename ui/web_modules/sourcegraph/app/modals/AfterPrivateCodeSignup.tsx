import * as React from "react";
import * as Relay from "react-relay";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import { Router, RouterLocation } from "sourcegraph/app/router";
import { Button, Input, Select } from "sourcegraph/components";
import { Heading } from "sourcegraph/components/index";
import { LocationStateModal, dismissModal } from "sourcegraph/components/Modal";
import * as modalStyles from "sourcegraph/components/styles/modal.css";
import { PopOut } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { langs } from "sourcegraph/Language";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { checkStatus, defaultFetch as fetch } from "sourcegraph/util/xhr";

interface GQLProps {
	relay: any;
	root: GQL.IRoot;
}

interface Props extends GQLProps {
	onSubmit?: () => void;
	style?: React.CSSProperties;
}

type State = {
	form: {
		fullName?: string;
		email?: string;
		language?: string;
		org?: string;
	};
};

class AfterPrivateCodeSignupForm extends React.Component<Props, State> {
	state: State = { form: {} };

	componentDidMount(): void {
		this._updateStateFromForm();
	}

	private _form: HTMLFormElement;

	_updateStateFromForm(): void {
		const formFields = this._form.elements;
		const newForm = {};
		for (let i = 0; i < formFields.length; i++) {
			const elem = formFields[i] as (HTMLInputElement | HTMLSelectElement);
			newForm[elem.name] = elem.value;
		}
		this.setState({ form: newForm });
	}

	_onChange(): void {
		this._updateStateFromForm();
	}

	_sendForm(ev: React.FormEvent<HTMLFormElement>): void {
		ev.preventDefault();

		let firstName = "";
		let lastName = "";
		if (this.state.form.fullName) {
			const names = this.state.form.fullName.split(/\s+/);
			firstName = names[0];
			lastName = names.slice(1).join(" ");
		}

		const hubspotProps = {
			firstname: firstName,
			lastname: lastName,
			email: this.state.form.email!,
			github_orgs: `,${this.state.form.org},`,
			is_personal_account_only: (this.state.form.org === context.user!.Login).toString(), // Go expects map[string]string, HubSpot auto-converts strings to booleans
			languages_used: this.state.form.language,
		};
		fetch(`/.api/submit-form`, {
			method: "POST",
			headers: { "Content-Type": "application/json; charset=utf-8" },
			body: JSON.stringify(hubspotProps),
		})
			.then(checkStatus)
			.then(() => {
				Events.AfterPrivateCodeSignup_Completed.logEvent({
					trialSignupProperties: hubspotProps,
				});

				if (this.props.onSubmit) {
					this.props.onSubmit();
				}
			})
			.catch(err => {
				if (this.props.onSubmit) {
					this.props.onSubmit();
				}
				throw new Error(`Submitting after signup form failed with error: ${err}`);
			});
	}

	render(): JSX.Element | null {
		let allOrgs: string[] = [];
		if (this.props.root && this.props.root.currentUser && this.props.root.currentUser.githubOrgs) {
			this.props.root.currentUser.githubOrgs.forEach(org => allOrgs.push(org));
		}
		if (context && context.user) {
			allOrgs.push(context.user.Login);
		}

		const isPersonalPlan = this.state.form.org === context.user!.Login;

		return (
			<div style={this.props.style}>
				<form onSubmit={ev => this._sendForm(ev)} onChange={ev => this._onChange()} ref={e => this._form = e}>
					<Input autoFocus={true} type="text" placeholder="Name" name="fullName" block={true} label="Your full name" containerStyle={{ marginBottom: whitespace[3] }} required={true} />
					<Select block={true} name="org" label="Your primary organization" containerSx={{ marginBottom: whitespace[3] }}>
						{allOrgs.map(org => <option value={org} key={org}>{org}{org === context.user!.Login ? " — personal account" : ""}</option>)}
					</Select>
					<p style={{ ...typography.size[6], color: colors.greenD1(), paddingBottom: whitespace[2] }}>{isPersonalPlan ? "Personal" : "Organization"}: 14 days free, then ${isPersonalPlan ? "9" : "25/user"}/month. Unlimited private repositories. <Link to="/pricing" target="_blank">Learn&nbsp;more&nbsp;<PopOut width={18} /></Link></p>
					<Input block={true} type="email" name="email" placeholder="you@example.com" label="Your work email" required={true} containerStyle={{ marginBottom: whitespace[3] }} />
					<Select name="language" containerSx={{ marginBottom: whitespace[3] }} label="Your primary programming language">
						{langs.map(([id, name]) => <option value={id} key={id}>{name}</option>)}
					</Select>
					<Button block={true} type="submit" color="purple" style={{ marginTop: whitespace[4] }}>Start using Sourcegraph</Button>
				</form>
			</div>
		);
	}
}

const Modal = (props: {
	location: RouterLocation;
	router: Router;
} & GQLProps): JSX.Element => {
	const sx = {
		maxWidth: "420px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return <LocationStateModal modalName="afterPrivateCodeSignup" location={props.location} router={props.router}>
		<div className={modalStyles.modal} style={sx}>
			<Heading level={4} align="center" underline="orange">Let&apos;s get started!</Heading>
			<AfterPrivateCodeSignupForm
				root={props.root}
				relay={props.relay}
				onSubmit={dismissModal("afterPrivateCodeSignup", props.location, props.router)} />
		</div>
	</LocationStateModal>;
};

const ModalContainer = Relay.createContainer(Modal, {
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				currentUser {
					githubOrgs
				}
			}`
	},
});

export const ModalMain = function (props: { location: RouterLocation, router: Router }): JSX.Element {
	if (!context || !context.user) {
		return <div />; // modal requires a user
	}
	return <Relay.RootContainer
		Component={ModalContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`query { root }`
			},
			params: props,
		}}
	/>;
};
