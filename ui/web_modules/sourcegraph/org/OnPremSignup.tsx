import * as React from "react";

import { Button, Heading, Input, TextArea } from "sourcegraph/components";
import { Postcard } from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";

export interface OnPremDetails {
	existingSoftware: string;
	versionControlSystem: string;
	numberOfDevs: string;
	otherDetails: string;
}

interface Props {
	next: (details: OnPremDetails) => void;
};

const formSx = {
	color: colors.blueGrayD1(),
	margin: whitespace[4],
	marginTop: 0,
};

export class EnterpriseDetails extends React.Component<Props, OnPremDetails> {
	state: OnPremDetails = {
		existingSoftware: "",
		versionControlSystem: "",
		numberOfDevs: "0",
		otherDetails: "",
	};

	updateField = (field: keyof OnPremDetails) => (ev: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const state = { ...this.state };
		state[field] = ev.currentTarget.value;
		this.setState(state);
	}

	submit = (ev: React.FormEvent<HTMLFormElement>) => {
		ev.preventDefault();
		this.props.next(this.state);
	}

	render(): JSX.Element {
		return <form style={formSx} onSubmit={this.submit}>
			<Input
				label="What software do you already use?"
				placeholder="Phabricator, CircleCI, etc."
				block={true}
				value={this.state.existingSoftware}
				onChange={this.updateField("existingSoftware")} />
			<Input
				label="What version control system do you use?"
				placeholder="Git, Subversion, etc."
				block={true}
				value={this.state.versionControlSystem}
				onChange={this.updateField("versionControlSystem")} />
			<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: whitespace[3] }}>
				<span>Number of developers</span>
				<Input
					value={this.state.numberOfDevs}
					type="number"
					containerStyle={{ marginBottom: 0, flex: "0 0 100px" }}
					onChange={this.updateField("numberOfDevs")} />
			</div>
			<TextArea
				label="Tell us more about your setup:"
				block={true}
				optionalText="Optional"
				style={{ height: 200 }}
				value={this.state.otherDetails}
				onChange={this.updateField("otherDetails")} />
			<div style={{ textAlign: "right" }}>
				<Button type="submit" color="blue" style={{ paddingLeft: whitespace[3], paddingRight: whitespace[3] }}>
					Submit
				</Button>
			</div>
		</form>;
	}
}

export function EnterpriseThanks(props: { next: () => void }): JSX.Element {
	return <div style={{ margin: "auto", maxWidth: 320, textAlign: "center", paddingTop: whitespace[4], paddingBottom: whitespace[5] }}>
		<Postcard width={64} color={colors.blueGrayL1()} />
		<Heading level={4}>Thanks!</Heading>
		<p>
			We'll get back to you as soon as possible, typically within 24 hours.
		</p>
		<p>
			In the meantime, enjoy browsing public code for free on Sourcegraph.
		</p>
		<Button onClick={props.next} color="blue" style={{ marginTop: whitespace[3] }}>
			Explore Sourcegraph
		</Button>
	</div>;
}
