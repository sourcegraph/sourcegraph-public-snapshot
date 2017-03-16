import * as format from "date-fns/format";
import * as React from "react";
import * as Relay from "react-relay";

import { context } from "sourcegraph/app/context";
import { Button, Heading } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { OrgPlan, PersonalPlan, PublicPlan } from "sourcegraph/components/PlanSelector";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { ChangeBillingInfo } from "sourcegraph/user/BillingInfo";
import { PlanChanger } from "sourcegraph/user/PlanChanger";
import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

interface Props {
	root: GQL.IRoot;
}

class Billing extends React.Component<Props, {}> {
	render(): JSX.Element {
		const user = this.props.root.currentUser!;
		return <div>
			<div style={{ padding: 40 }}>
				<PageTitle title="Billing information" />
				<div style={{
					display: "flex",
					alignItems: "center",
					justifyContent: "space-between",
					marginBottom: 24,
				}}>
					<Heading level={6}>Your plan</Heading>
					<PlanChanger />
				</div>
				<PlanTile plan={user.paymentPlan} org={user.paymentPlan.organization} />
			</div>
			<BillingDetails plan={user.paymentPlan} />
		</div>;
	}
}

function formatRenewalDate(d: Date): string {
	return format(d, "MMMM Do YYYY");
}

const modalName = "cancelSubscriptionModal";

class CancelSubscription extends ComponentWithRouter<{ plan: GQL.IPlan }, {}> {

	private cancelSubcription = () => {
		Events.CancelSubscription_Clicked.logEvent();
		fetchGraphQLQuery(`mutation {
				cancelSubscription()
			}`);
		location.reload();
	}

	render(): JSX.Element {
		const date = formatRenewalDate(new Date(this.props.plan.renewalDate! * 1000));
		return <div>
			<LocationStateToggleLink modalName={modalName} location={this.context.router.location} onToggle={v => v && Events.CancelSubscriptionModal_Initiated.logEvent()}>
				Disable auto-renewal
			</LocationStateToggleLink>
			<LocationStateModal style={{ textAlign: "center" }} modalName={modalName} title="Confirm cancelation">
				Are you sure you want to disable auto renewal? Your
				subscription will end on {date}.
				<div style={{ marginTop: 32 }}>
					<Button onClick={this.cancelSubcription}>Confirm</Button>
				</div>
			</LocationStateModal>
		</div>;
	}
}

function BillingDetails({ plan }: { plan: GQL.IPlan }): JSX.Element {
	if (plan.name !== "organization") {
		return <div />;
	}
	const date = formatRenewalDate(new Date(plan.renewalDate! * 1000));
	return <div>
		<hr />
		<div style={{ padding: 40 }}>
			<div style={{
				display: "flex",
				alignItems: "center",
				justifyContent: "space-between",
				marginBottom: 24,
			}}>
				<Heading level={6}>Billing information</Heading>
				<ChangeBillingInfo />
			</div>
			Your annual subscription will renew on {date} for ${plan.cost! / 100}.
			<CancelSubscription plan={plan} />
		</div>
	</div>;
}

function PlanTile({ plan, org }: { plan: GQL.IPlan, org: GQL.IOrganization | null }): JSX.Element {
	if (plan.name === "private") {
		return <div>
			<PersonalPlan />
			Your plan allows you to view code hosted under your account on GitHub.
		</div>;
	} else if (plan.name === "organization") {
		if (!org) {
			throw new Error("Expected organization.");
		}
		return <div>
			<OrgPlan />
			Your plan allows {plan.seats} {plan.seats === 1 ? "person" : "people"} to view code from the {org.name} organization.
		</div>;
	}
	return <PublicPlan />;
}

export class BillingContainer extends ComponentWithRouter<{}, {}> {
	container: Relay.RelayContainerClass<Billing> = Relay.createContainer(Billing, {
		fragments: {
			root: () => Relay.QL`
				fragment on Root {
					currentUser {
						paymentPlan {
							seats
							name
							cost
							renewalDate
							organization {
								name
								avatarURL
							}
						}
					}
				}`,
		}
	});

	render(): JSX.Element {
		if (!context || !context.user) {
			return <div>
				Please <LocationStateToggleLink href="/login" modalName="login" location={this.context.router.location} onToggle={v => v && Events.LoginModal_Initiated.logEvent({ page_name: location.pathname })}>
					log in
				</LocationStateToggleLink> to view this page.
			</div>;
		}
		return <Relay.RootContainer
			Component={this.container}
			route={{
				name: "Root",
				queries: {
					root: () => Relay.QL`query { root }`
				},
				params: {},
			}}
		/>;
	}
}
