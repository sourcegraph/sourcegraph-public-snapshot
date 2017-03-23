import * as React from "react";
import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
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

interface FormData {
	fullName: string;
	email: string;
	company: string;
	editors: string[];
	languages: string[];
	message: string;
}

interface State {
	formError?: string;
	resp?: Response;
	respError?: Error;
	formData: FormData;
};

export class BetaInterestForm extends React.Component<Props, State> {
	state: State = { formData: { fullName: "", email: "", company: "", editors: [], languages: [], message: "", } };

	componentWillMount(): void {
		const formData = this.loadStateFromLocalStorage();
		if (formData) {
			this.state.formData = formData;
		}
	}

	componentDidUpdate(): void {
		this.saveStateToLocalStorage();
	}

	private saveStateToLocalStorage = (): void => {
		const formData = this.state.formData;
		localStorage.setItem("beta-interest-form", JSON.stringify({
			fullName: formData.fullName,
			email: formData.email,
			company: formData.company,
			editors: formData.editors,
			languages: formData.languages,
			message: formData.message,
		}));
	}

	private loadStateFromLocalStorage = (): FormData | null => {
		const ls = localStorage.getItem("beta-interest-form");
		if (ls) {
			const lsParsed = JSON.parse(ls);
			return {
				fullName: lsParsed.fullName || "",
				email: lsParsed.email || "",
				company: lsParsed.company || "",
				editors: lsParsed.editors || [],
				languages: lsParsed.languages || [],
				message: lsParsed.message || "",
			};
		}
		return null;
	}

	private onInputChange = (field: keyof FormData) => (ev: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const state = { ...this.state, formData: { ...this.state.formData } };
		state.formData[field] = ev.currentTarget.value;
		this.setState(state);
	}

	private onCheckboxChange = (field: keyof FormData) => (list: string[]) => {
		const state = { ...this.state, formData: { ...this.state.formData } };
		state.formData[field] = list;
		this.setState(state);
	}

	private sendForm = (ev: any): void => {
		ev.preventDefault();
		const formData = this.state.formData;
		let firstName;
		let lastName;
		if (formData.fullName) {
			const names = formData.fullName.split(/\s+/);
			firstName = names[0];
			lastName = names.slice(1).join(" ");
		}

		if (formData.editors.length === 0) {
			this.setState({ ...this.state, formError: "Please select at least one preferred editor." });
			return;
		}
		if (formData.languages.length === 0) {
			this.setState({ ...this.state, formError: "Please select at least one preferred language." });
			return;
		}
		submitBetaSignupForm({
			email: formData.email,
			firstname: firstName || "",
			lastname: lastName || "",
			company: formData.company,
			languages_used: formData.languages,
			editors_used: formData.editors,
			message: formData.message.trim(),
		}).then(resp => {
			this.setState({ ...this.state, resp });
		}).catch(respError => {
			this.setState({ ...this.state, respError });
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
		let defaults = this.state.formData;

		if (language) {
			defaults.languages.push(langName(language));
		}

		return (
			<div style={this.props.style}>
				{betaRegistered && <span>
					<p>You've already registered. We'll contact you once a beta matching your interests has begun.</p>
					<p>Feel free to update your favorite editors / languages using the form below.</p>
				</span>}
				<form className={className} onSubmit={this.sendForm}>
					<Input
						block={true}
						type="text"
						name="fullName"
						placeholder="Name"
						required={true}
						defaultValue={defaults.fullName}
						onChange={this.onInputChange("fullName")} />
					{(!emails || emails.length === 0) &&
						<Input
							block={true}
							type="email"
							name="email"
							placeholder="Email address"
							required={true} defaultValue={defaults.email}
							onChange={this.onInputChange("email")} />
					}
					<Input
						block={true}
						type="text"
						name="company"
						placeholder="Company / organization"
						required={true}
						defaultValue={defaults.company}
						onChange={this.onInputChange("company")} />
					<CheckboxList
						title="Preferred editors"
						name="editors" labels={editors}
						defaultValues={defaults.editors}
						style={{ marginBottom: whitespace[3] }}
						onChange={this.onCheckboxChange("editors")} />
					<CheckboxList
						title="Preferred languages"
						name="languages"
						labels={languageNames}
						values={languageIDs}
						defaultValues={defaults.languages}
						style={{ marginBottom: whitespace[3] }}
						onChange={this.onCheckboxChange("languages")} />
					<TextArea
						block={true}
						name="message"
						placeholder="Other / comments"
						defaultValue={defaults.message}
						onChange={this.onInputChange("message")}>
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
