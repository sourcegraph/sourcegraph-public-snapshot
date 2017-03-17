import * as React from "react";
import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
import { Component } from "sourcegraph/Component";
import { Button, CheckboxList, Input, Panel, TextArea } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";
import { editors, languageIDs, languageNames } from "sourcegraph/home/HomeUtils";
import { langName } from "sourcegraph/Language";
import { SignupForm } from "sourcegraph/user/Signup";
import { submitBetaSignupForm } from "sourcegraph/user/SubmitForm";

interface Props {
	onSubmit?: () => void;
	className?: string;
	language?: string;
	location: RouterLocation;
	style?: React.CSSProperties;
}

interface State {
	formError?: string;
	resp?: Response;
	respError?: Error;
};

export class BetaInterestForm extends Component<Props, State> {
	private fullName: HTMLInputElement;
	private email: HTMLInputElement;
	private company: HTMLInputElement;
	private editors: CheckboxList;
	private languages: CheckboxList;
	private message: HTMLTextAreaElement;

	componentDidMount(): void {
		// Trigger onChange now to save this.props.language if set.
		if (context.user && this.props.language) {
			this.onChange();
		}
	}

	private onChange = (): void => {
		localStorage.setItem("beta-interest-form", JSON.stringify({
			fullName: this.fullName.value,
			email: this.email ? this.email.value : "",
			company: this.company.value,
			editors: this.editors.selected(),
			languages: this.languages.selected(),
			message: this.message.value,
		}));
	}

	private sendForm = (ev: any): void => {
		ev.preventDefault();
		const name = this.fullName.value;
		let firstName;
		let lastName;
		if (name) {
			const names = name.split(/\s+/);
			firstName = names[0];
			lastName = names.slice(1).join(" ");
		}

		if (this.editors.selected().length === 0) {
			this.setState({ formError: "Please select at least one preferred editor." });
			return;
		}
		if (this.languages.selected().length === 0) {
			this.setState({ formError: "Please select at least one preferred language." });
			return;
		}

		submitBetaSignupForm({
			email: this.email ? this.email.value.trim() : "",
			firstname: firstName || "",
			lastname: lastName || "",
			company: this.company.value,
			languages_used: this.languages.selected(),
			editors_used: this.editors.selected(),
			message: this.message.value.trim(),
		}).then(resp => {
			this.setState({ resp });
		}).catch(respError => {
			this.setState({ respError });
		});
	}

	render(): JSX.Element | null {
		if (this.state.resp && !this.state.respError) {
			// Display a "Close" button if there is an onSubmit handler.
			return (<span>
				<p>Success! Return to this page any time to update your favorite editors / languages!</p>
				{this.state.resp["EmailAddress"] !== undefined
					? <p>We'll contact you at <strong>{this.state.resp["EmailAddress"]}</strong> once a beta has begun.</p>
					: <p>We'll contact you once a beta has begun.</p>}
				{this.props.onSubmit && <Button block={true} type="submit" color="purple" onClick={this.props.onSubmit}>Close</Button>}
			</span>);
		}

		if (!context.user) {
			return <Panel hoverLevel="low" style={{ paddingTop: whitespace[4], textAlign: "center" }}>
				<p>You must sign up to continue.</p>
				<SignupForm />
			</Panel>;
		}

		let [className, language] = [this.props.className, this.props.language];
		let betaRegistered = false; // TODO
		let emails = context.emails && context.emails.EmailAddrs;

		let defaultFullName;
		let defaultEmail;
		let defaultCompany;
		let defaultMessage;
		let defaultEditors = [];
		let defaultLanguages: string[] = [];
		const ls = localStorage.getItem("beta-interest-form");
		if (ls) {
			const lsParsed = JSON.parse(ls);
			defaultFullName = lsParsed.fullName;
			defaultEmail = lsParsed.email;
			defaultCompany = lsParsed.company;
			defaultEditors = lsParsed.editors;
			defaultLanguages = lsParsed.languages;
			defaultMessage = lsParsed.message;
		}

		if (language) {
			defaultLanguages.push(langName(language));
		}

		return (
			<div style={this.props.style}>
				{betaRegistered && <span>
					<p>You've already registered. We'll contact you once a beta matching your interests has begun.</p>
					<p>Feel free to update your favorite editors / languages using the form below.</p>
				</span>}
				<form className={className} onSubmit={this.sendForm} onChange={this.onChange}>
					<Input
						domRef={c => this.fullName = c}
						block={true}
						type="text"
						name="fullName"
						placeholder="Name"
						required={true}
						defaultValue={defaultFullName} />
					{(!emails || emails.length === 0) &&
						<Input
							domRef={c => this.email = c}
							block={true}
							type="email"
							name="email"
							placeholder="Email address"
							required={true} defaultValue={defaultEmail} />
					}
					<Input
						domRef={c => this.company = c}
						block={true}
						type="text"
						name="company"
						placeholder="Company / organization"
						required={true}
						defaultValue={defaultCompany} />
					<CheckboxList
						ref={c => this.editors = c}
						title="Preferred editors"
						name="editors" labels={editors}
						defaultValues={defaultEditors}
						style={{ marginBottom: whitespace[3] }} />
					<CheckboxList
						ref={c => this.languages = c}
						title="Preferred languages"
						name="languages"
						labels={languageNames}
						values={languageIDs}
						defaultValues={defaultLanguages}
						style={{ marginBottom: whitespace[3] }} />
					<TextArea
						block={true}
						domRef={c => this.message = c}
						name="message"
						placeholder="Other / comments"
						defaultValue={defaultMessage}>
					</TextArea>
					<Button block={true} type="submit" color="purple">
						{betaRegistered ? "Update my interests" : "Participate in the beta"}
					</Button>
					<div className={base.pb4}>
						<br />
						{this.state.formError && <strong>{this.state.formError}</strong>}
						{this.state.respError && <div>{this.state.respError.message}</div>}
					</div>
				</form>
			</div>
		);
	}
}
