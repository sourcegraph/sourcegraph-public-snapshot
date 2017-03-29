import * as React from "react";
import { RouterLocation } from "sourcegraph/app/router";
import { Button, CheckboxList, Input, TextArea } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";
import { editors } from "sourcegraph/home/HomeUtils";
import { submitZapBetaSignupForm } from "sourcegraph/user/SubmitForm";

interface Props {
	onSubmit?: () => void;
	className?: string;
	language?: string;
	location: RouterLocation;
	style?: React.CSSProperties;
}

interface FormData {
	firstName: string;
	lastName: string;
	email: string;
	company: string;
	editors: string[];
	message: string;
}

interface State {
	formError?: string;
	resp?: Response;
	respError?: Error;
	formData: FormData;
};

export class ZapBetaInterestForm extends React.Component<Props, State> {
	state: State = { formData: { firstName: "", lastName: "", email: "", company: "", editors: [], message: "", } };

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
		localStorage.setItem("zap-beta-interest-form", JSON.stringify({
			firstName: formData.firstName,
			lastName: formData.lastName,
			email: formData.email,
			company: formData.company,
			editors: formData.editors,
			message: formData.message,
		}));
	}

	private loadStateFromLocalStorage = (): FormData | null => {
		const ls = localStorage.getItem("zap-beta-interest-form");
		if (ls) {
			const lsParsed = JSON.parse(ls);
			return {
				firstName: lsParsed.firstName || "",
				lastName: lsParsed.lastName || "",
				email: lsParsed.email || "",
				company: lsParsed.company || "",
				editors: lsParsed.editors || [],
				message: lsParsed.message.trim() || "",
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

		if (formData.editors.length === 0) {
			this.setState({ ...this.state, formError: "Please select at least one preferred editor." });
			return;
		}

		submitZapBetaSignupForm({
			beta_email: formData.email,
			firstname: formData.firstName,
			lastname: formData.lastName,
			company: formData.company,
			editors_used: formData.editors,
			message: formData.message,
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
				<hr />
				<p>Success! Return to this page any time to update your favorite editors!</p>
				{this.state.resp["EmailAddress"] !== undefined
					? <p>We'll contact you at <strong>{this.state.resp["EmailAddress"]}</strong> once a beta has begun.</p>
					: <p>We'll contact you once a beta has begun.</p>}
				{this.props.onSubmit && <Button block={true} type="submit" color="purple" onClick={this.props.onSubmit}>Close</Button>}
			</span>);
		}

		let className = this.props.className;
		let defaults = this.state.formData;

		return (
			<div style={this.props.style}>
				<form className={className} onSubmit={this.sendForm}>
					<Input
						block={true}
						type="text"
						name="firstName"
						placeholder="First name"
						required={true}
						defaultValue={defaults.firstName}
						onChange={this.onInputChange("firstName")} />
					<Input
						block={true}
						type="text"
						name="lastName"
						placeholder="Last name"
						required={true}
						defaultValue={defaults.lastName}
						onChange={this.onInputChange("lastName")} />
					<Input
						block={true}
						type="email"
						name="email"
						placeholder="Email address"
						required={true}
						defaultValue={defaults.email}
						onChange={this.onInputChange("email")} />
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
					<TextArea
						block={true}
						name="message"
						placeholder="Other editors / comments"
						defaultValue={defaults.message}
						onChange={this.onInputChange("message")}>
					</TextArea>
					<Button block={true} type="submit" color="purple">
						Participate in the beta
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
