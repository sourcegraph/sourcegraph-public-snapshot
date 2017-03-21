import * as React from "react";
import { Link } from "react-router";

import { abs } from "sourcegraph/app/routePatterns";
import { Button, FlexContainer, Input } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

const detailsSx = {
	padding: "2rem 50px",
};

export interface UserDetails {
	name: string;
	email: string;
	company: string;
}

interface Props {
	next: (userDetails: UserDetails) => void;
}

export class UserDetailsForm extends React.Component<Props, UserDetails> {
	state: UserDetails = {
		name: "",
		email: "",
		company: "",
	};

	private submit = (ev: React.FormEvent<HTMLFormElement>) => {
		ev.preventDefault();
		this.props.next(this.state);
	}

	private onChange = (field: keyof UserDetails) => (ev: React.FormEvent<HTMLInputElement>) => {
		const state = { ...this.state };
		state[field] = ev.currentTarget.value;
		this.setState(state);
	}

	render(): JSX.Element {
		const subtleLinkSx = {
			color: colors.blueGrayD1(),
		};

		return <form style={detailsSx} onSubmit={this.submit}>
			<p style={{ marginTop: 0, marginBottom: whitespace[5], textAlign: "center" }}>
				Please enter your details:
				</p>
			<Input label="Full name" block={true} required={true} value={this.state.name} onChange={this.onChange("name")} />
			<Input label="Email" block={true} required={true} value={this.state.email} onChange={this.onChange("email")} type="email" />
			<Input block={true} label="Company" required={false} optionalText="Optional" value={this.state.company} onChange={this.onChange("company")} />

			<FlexContainer style={{ marginBottom: whitespace[3], marginTop: whitespace[5] }}>
				<div style={{ color: colors.blueGray(), ...typography.small }}>
					By signing up, you agree to our <Link to={abs.terms} style={subtleLinkSx}>terms</Link>, <Link to={abs.privacy} style={subtleLinkSx}>privacy policy</Link>, and <Link to={abs.terms} style={subtleLinkSx}>email policy</Link>.
					</div>
				<Button type="submit" color="blue" size="small" style={{ flex: "0 0 120px" }}>Sign up</Button>
			</FlexContainer>
		</form >;
	}
}
