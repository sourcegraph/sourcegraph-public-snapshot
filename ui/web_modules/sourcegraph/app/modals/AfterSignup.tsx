import * as React from "react";
import * as Relay from "react-relay";

import { context } from "sourcegraph/app/context";
import { Router, RouterLocation } from "sourcegraph/app/router";
import { LocationStateModal, dismissModal } from "sourcegraph/components/Modal";
import { PlanSelector, PlanType } from "sourcegraph/components/PlanSelector";
import { EnterpriseDetails, EnterpriseThanks, OnPremDetails } from "sourcegraph/org/OnPremSignup";
import { OrgSelection } from "sourcegraph/org/OrgSignup";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";
import { submitAfterSignupForm } from "sourcegraph/user/SubmitForm";
import { UserDetails, UserDetailsForm } from "sourcegraph/user/UserDetails";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

interface Props {
	onSubmit?: () => void;
	root: GQL.IRoot;
}

type Stage = "details" | "plan" | "enterpriseDetails" | "orgDetails" | "enterpriseThanks" | "finished";

interface Details {
	stage: Stage;
	authedPrivate: boolean;
	organization?: string;
	onPremDetails?: OnPremDetails;
	plan?: PlanType;
	userInfo?: UserDetails;
};

function submitSignupInfo(details: Details): void {
	if (!details.userInfo) {
		throw new Error("Expected user info to be filled out");
	}
	let firstName = "";
	let lastName = "";
	if (details.userInfo.name) {
		const names = details.userInfo.name.split(/\s+/);
		firstName = names[0];
		lastName = names.slice(1).join(" ");
	}

	let signupInformation = {
		firstname: firstName,
		lastname: lastName,
		company: details.userInfo.company,
		signupEmail: details.userInfo.email,
		planOrgs: details.organization,
		plan: details.plan === undefined ? "public" : details.plan,
		isPrivateCodeUser: JSON.stringify(details.authedPrivate),
	};

	if (details.onPremDetails) {
		signupInformation = {
			...signupInformation,
			...details.onPremDetails,
		};
	}

	if (details.plan === "organization") {
		fetchGraphQLQuery(`mutation {
			startOrgTrial(GitHubOrg: $org)
		}`, { org: details.organization });
	}

	Events.AfterSignupModal_Completed.logEvent({
		trialSignupProperties: signupInformation,
	});
	EventLogger.setUserRegisteredAt(Date.now());

	submitAfterSignupForm(signupInformation);
}

export class AfterSignupForm extends React.Component<Props, Details> {

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	authedPrivate: boolean = this.context.router.location.query["private"] === "true";

	state: Details = {
		stage: "details",
		authedPrivate: this.authedPrivate,
	};

	private submit = () => {
		submitSignupInfo(this.state);
		if (this.props.onSubmit) {
			this.props.onSubmit();
		}
	}

	private selectPlan = (plan: PlanType) => () => {
		Events.SignupPlan_Selected.logEvent({
			signup: { plan }
		});
		EventLogger.setUserPlan(plan);
		let stage;
		if (plan === "enterprise") {
			stage = "enterpriseDetails";
		} else if (plan === "organization") {
			stage = "orgDetails";
		} else {
			stage = "finished";
		}
		this.setState({ ...this.state, plan, stage });
	}

	private gotoPlans = () => {
		this.setState({ ...this.state, stage: "plan" });
	}

	private selectOrg = (organization: string) => () => {
		Events.SignupOrg_Selected.logEvent({
			signup: { organization },
		});
		EventLogger.setUserPlanOrg(organization);
		this.setState({ ...this.state, stage: "finished", organization });
	}

	private onPremComplete = (onPremDetails: OnPremDetails) => {
		Events.SignupEnterpriseForm_Completed.logEvent({
			signup: { onPremDetails: onPremDetails },
		});
		this.setState({ ...this.state, onPremDetails, stage: "enterpriseThanks" });
	}

	private userDetailsComplete = (userInfo: UserDetails) => {
		Events.SignupUserDetails_Completed.logEvent({
			signup: { userInfo: userInfo },
		});
		EventLogger.setUserName(userInfo.name);
		if (userInfo.company.length > 0) {
			EventLogger.setUserCompany(userInfo.company);
		}
		let stage;
		if (this.authedPrivate) {
			stage = "plan";
		} else {
			stage = "finished";
			// If user did not auth private code, set user prop `plan` to be public
			EventLogger.setUserPlan("public");
		}
		this.setState({ ...this.state, stage, userInfo });
	}

	private logStage(stage: Stage): void {
		Events.SignupStage_Initiated.logEvent({
			signup: { stage },
		});
	}

	componentWillUpdate(_: Props, state: Details): void {
		if (state.stage !== this.state.stage) {
			this.logStage(state.stage);
		}
	}

	componentDidMount(): void {
		this.logStage(this.state.stage);
	}

	componentDidUpdate(): void {
		if (this.state.stage === "finished") {
			this.submit();
		}
	}

	render(): JSX.Element | null {
		switch (this.state.stage) {
			case "details":
				return <UserDetailsForm next={this.userDetailsComplete} />;
			case "plan":
				return <PlanSelector select={this.selectPlan} />;
			case "enterpriseDetails":
				return <EnterpriseDetails next={this.onPremComplete} />;
			case "orgDetails":
				return <OrgSelection root={this.props.root} back={this.gotoPlans} select={this.selectOrg} />;
			case "enterpriseThanks":
				return <EnterpriseThanks next={this.submit} />;
			default:
				return null;
		}
	}
}

const Modal = (props: {
	location: RouterLocation;
	router: Router;
	root: GQL.IRoot;
}): JSX.Element => {
	return <LocationStateModal modalName="afterSignup" title="Sign up" padded={false} sticky={true}>
		<AfterSignupForm
			root={props.root}
			onSubmit={dismissModal("afterSignup", props.router)} />
	</LocationStateModal>;
};

const ModalContainer = Relay.createContainer(Modal, {
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				currentUser {
					githubOrgs {
						name
						avatarURL
						description
						collaborators
					}
				}
			}`
	},
});

export class ModalMain extends React.Component<{}, {}> {

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	render(): JSX.Element {
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
				params: {
					router: this.context.router,
					location: this.context.router.location
				},
			}}
		/>;
	}
};
