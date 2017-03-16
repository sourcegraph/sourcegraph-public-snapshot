import * as React from "react";
import * as Relay from "react-relay";

import { context } from "sourcegraph/app/context";
import { Button, FlexContainer, Toast } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { OrgSeatsCard } from "sourcegraph/components/OrganizationCard";
import { ChevronRight, No, Warning } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";
import { PaymentInfo } from "sourcegraph/user/BillingInfo";
import { PlanChanger } from "sourcegraph/user/PlanChanger";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

const sales = <a href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>;

export function Paywall({ repo }: { repo: GQL.IRepository }): JSX.Element {
	EventLogger.logViewEvent("ViewPaywall", location.pathname, {});
	return <FlexContainer direction="top-bottom" style={{ flex: "1 1 auto" }}>
		<Toast isDismissable={false} color="gray" style={{ flex: "0 0 auto" }}>
			<div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
				<div>
					<No color={colors.orangeL1()} width={24} style={{ marginRight: whitespace[3] }} />
					Your trial has ended.
			</div>
				<CompleteTrialContainer repo={repo} />
			</div>
		</Toast>
		<div style={{
			backgroundColor: colors.blueGrayD2(),
			backgroundImage: `url(${context.assetsRoot}/img/blur-screenshot.png)`,
			backgroundSize: "cover",
			backgroundRepeat: "no-repeat",
			flex: "1 1 auto",
		}} />
	</FlexContainer>;
}

export function TrialEndingWarning({ layout, repo }: {
	layout: () => void, repo: GQL.IRepository
}): JSX.Element {
	if (!repo.expirationDate) {
		return <div />;
	}
	const time = new Date(repo.expirationDate * 1000);
	const msUntilExpiration = time.getTime() - Date.now();
	const fiveDays = 1000 * 60 * 60 * 24 * 5;
	if (msUntilExpiration > fiveDays || msUntilExpiration < 0) {
		return <div />;
	}
	const daysLeft = new Date(msUntilExpiration).getUTCDate() - 1;
	EventLogger.logViewEvent("ViewTrialExpirationBanner", location.pathname, { daysLeft });
	let timeLeft: string;
	switch (daysLeft) {
		case 0: timeLeft = "today"; break;
		case 1: timeLeft = "tomorrow"; break;
		case 2: timeLeft = "in two days"; break;
		case 3: timeLeft = "in three days"; break;
		case 4: timeLeft = "in four days"; break;
		case 5: timeLeft = "in five days"; break;
		default: timeLeft = "soon";
	}
	return <Toast color="gray" isDismissable={true} style={{ zIndex: 6 }}>
		<Warning color={colors.yellow()} width={24} style={{ marginRight: whitespace[3] }} />
		Your free trial is ending {timeLeft}. Please contact {sales} to continue using
		Sourcegraph on private code.
	</Toast>;
}

const SEAT_COST = 300;

interface State {
	seats: string;
	showPayment: boolean;
}

const modalName = "trialCompletionModal";

class CompleteTrialButton extends ComponentWithRouter<Props & { root: GQL.IRoot }, State> {

	state: State = {
		seats: "1",
		showPayment: false,
	};

	paymentClosed = () => {
		this.setState({ ...this.state, showPayment: false });
	}

	showStripe = () => {
		const seats = parseInt(this.state.seats, 10);
		if (seats < 1) {
			return;
		}
		this.setState({ ...this.state, showPayment: true });
	}

	onChange = (ev: React.FormEvent<HTMLInputElement>) => {
		this.setState({ ...this.state, seats: ev.currentTarget.value });
	}

	submitPayment = (token: any) => {
		const org = this.getOrg();
		if (!org) {
			throw new Error("Must have organization to purchase plan");
		}
		const seats = parseInt(this.state.seats, 10);
		if (seats < 1) {
			return;
		}
		fetchGraphQLQuery(`mutation {
			subscribeOrg(tokenID: $tokenID, GitHubOrg: $GitHubOrg, seats: $seats)
		}`, {
				tokenID: token.id,
				seats,
				GitHubOrg: org.name,
			});
		location.reload();
	}

	getOrg(): GQL.IOrganization | null {
		const orgs = this.props.root.currentUser!.githubOrgs;
		const match = this.props.repo.uri.match(/^.*\/(.*)\/.*$/);
		if (!match) {
			return null;
		}
		const orgName = match[1];
		for (const org of orgs) {
			if (org.name === orgName) {
				return org;
			}
		}
		return null;
	}

	render(): JSX.Element {
		const seats = parseInt(this.state.seats, 10);
		const cost = isNaN(seats) ? 0 : SEAT_COST * seats;
		const organization = this.getOrg();
		if (!organization) {
			return <Button>
				<PlanChanger />
			</Button>;
		}
		return <div>
			<LocationStateToggleLink modalName={modalName} location={this.context.router.location} onToggle={v => v && Events.CompleteSubscriptionModal_Initiated.logEvent()}>
				<Button color="blue">Complete your subscription <ChevronRight /></Button>
			</LocationStateToggleLink>
			{this.state.showPayment ?
				<PaymentInfo
					submit={this.submitPayment}
					closed={this.paymentClosed}
					description="1 year organization subscription"
					amount={cost * 100} /> :
				<LocationStateModal
					style={{
						textAlign: "center",
						color: colors.blueGrayD1(),
					}}
					modalName={modalName}
					padded={false}
					title="Complete your subscription">
					<div style={{ padding: 32 }}>
						<div>Confirm your organization and the number of seats you wish to purchase.</div>
						<div style={{
							marginTop: 24,
							color: colors.blueGray(),
							...typography.small
						}}>
							You will be billed <b>${cost}</b> and yearly thereafter.
						<OrgSeatsCard seats={this.state.seats} org={organization} onChange={this.onChange} />
						</div>
					</div>
					<hr />
					<div style={{ padding: 32, display: "flex", justifyContent: "space-between" }}>
						<div style={{ textAlign: "left" }}>
							<b>${cost} per year</b><br />
							{this.state.seats} {seats === 1 ? "seat" : "seats"} for 1 organization
					</div>
						<Button color="blue" onClick={this.showStripe}>
							Check out
					</Button>
					</div>
				</LocationStateModal>}
		</div >;
	}
}

export function needsPayment(repo: GQL.IRepository): boolean {
	if (!repo.expirationDate) {
		return false;
	}
	const time = new Date(repo.expirationDate * 1000);
	const msUntilExpiration = time.getTime() - Date.now();
	return msUntilExpiration < 0;
}

interface Props {
	repo: GQL.IRepository;
}

class CompleteTrialContainer extends React.Component<Props, {}> {
	container: Relay.RelayContainerClass<CompleteTrialButton> = Relay.createContainer(CompleteTrialButton, {
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
				}
			`,
		}
	});

	render(): JSX.Element {
		return <Relay.RootContainer
			Component={this.container}
			route={{
				name: "Root",
				queries: {
					root: () => Relay.QL`
					query { root }
				`,
				},
				params: {
					repo: this.props.repo,
				},
			}}
		/>;
	}
}
