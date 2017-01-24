import * as React from "react";
import * as Relay from "react-relay";

import { context } from "sourcegraph/app/context";
import { Router, RouterLocation } from "sourcegraph/app/router";
import { Button, Input, Select } from "sourcegraph/components";
import { Heading } from "sourcegraph/components/index";
import { LocationStateModal, dismissModal } from "sourcegraph/components/Modal";
import * as modalStyles from "sourcegraph/components/styles/modal.css";
import { whitespace } from "sourcegraph/components/utils";
import { langs } from "sourcegraph/Language";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
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
		org?: string;
		language?: string;
	};
};

class AfterPrivateCodeSignupForm extends React.Component<Props, State> {
	_onChange(ev: React.FormEvent<HTMLFormElement>): void {
		const el: HTMLInputElement = ev.target as any;
		this.setState({
			form: { ...this.state.form, [el.name]: el.value },
		});
	}

	_sendForm(ev: React.FormEvent<HTMLFormElement>): void {
		ev.preventDefault();

		const form: {
			fullName?: string;
			email?: string;
			language?: string;
			org?: string;
		} = {};
		const formFields = (ev.target as HTMLFormElement).elements;
		for (let i = 0; i < formFields.length; i++) {
			const elem = formFields[i] as (HTMLInputElement | HTMLSelectElement);
			form[elem.name] = elem.value;
		}

		let firstName;
		let lastName;
		if (form.fullName) {
			const names = form.fullName.split(/\s+/);
			firstName = names[0];
			lastName = names.slice(1).join(" ");
		}

		fetch(`/.api/submit-form`, {
			method: "POST",
			headers: { "Content-Type": "application/json; charset=utf-8" },
			body: JSON.stringify({
				firstname: firstName,
				lastname: lastName,
				email: form.email!,
				company: form.org!,
				languages_used: form.language!,
			}),
		})
			.then(checkStatus)
			.catch((err) => {
				console.error("Error submitting form:", err); // tslint:disable-line no-console
			})
			.then(() => {
				AnalyticsConstants.Events.AfterPrivateCodeSignup_Completed.logEvent({
					trialSignupProperties: {
						firstname: firstName,
						lastname: lastName,
						email: form.email,
						company: form.org,
						languages_used: form.language,
					}
				});

				if (this.props.onSubmit) {
					this.props.onSubmit();
				}
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

		return (
			<div style={this.props.style}>
				<form onSubmit={ev => this._sendForm(ev)}>
					<Input autoFocus={true} type="text" placeholder="Name" name="fullName" block={true} label="Your full name" containerStyle={{ marginBottom: whitespace[3] }} required={true} />
					<Select block={true} name="org" label="Your organization" containerSx={{ marginBottom: whitespace[3] }}>
						{allOrgs.map(org => <option value={org} key={org}>{org}</option>)}
					</Select>
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
