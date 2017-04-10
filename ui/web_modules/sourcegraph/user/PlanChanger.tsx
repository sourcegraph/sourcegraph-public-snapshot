import * as React from "react";

import { Button, Heading, TextArea } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { submitChangeUserPlanForm } from "sourcegraph/user/SubmitForm";

interface State {
	contents: string;
	active: boolean;
	submitted: boolean;
}

const modalName = "planChanger";

export class PlanChanger extends ComponentWithRouter<{}, State> {
	state: State = {
		contents: "",
		active: false,
		submitted: false,
	};

	contentsUpdated = (ev: React.FormEvent<HTMLTextAreaElement>) => {
		this.setState({ ...this.state, contents: ev.currentTarget.value });
	}

	submit = () => {
		submitChangeUserPlanForm({
			changePlanMessage: this.state.contents,
		});
		this.setState({ ...this.state, submitted: true });
	}

	modalDismissed = () => {
		this.setState({ ...this.state, submitted: false, contents: "" });
	}

	showModal(): void {
		Events.ChangePlanModal_Initiated.logEvent();
		this.forceUpdate();
	}

	render(): JSX.Element {
		return <div>
			<LocationStateToggleLink location={this.context.router.location} modalName={modalName} onToggle={v => v && this.showModal()}>
				Change your plan
			</LocationStateToggleLink>
			<LocationStateModal title="Change your plan" modalName={modalName} onDismiss={this.modalDismissed}>
				{this.state.submitted ?
					<div style={{ textAlign: "center" }}>
						<Heading level={3}>Thanks</Heading>
						We'll update your account as soon as possible, typically with 24 hours. Please <a
							href="mailto:hi@sourcegraph.com">contact us</a> if
						you	have any questions or concerns.
					</div> :
					<div>
						Describe the changes you'd like to make to your plan:
						<TextArea
							placeholder="Number of seats, attached organization"
							style={{ height: 400, marginTop: 16 }}
							block={true}
							value={this.state.contents}
							onChange={this.contentsUpdated} />
						<div style={{ textAlign: "right" }}>
							<Button color="blue" onClick={this.submit}>Submit</Button>
						</div>
					</div>}
			</LocationStateModal>
		</div>;
	}
}
