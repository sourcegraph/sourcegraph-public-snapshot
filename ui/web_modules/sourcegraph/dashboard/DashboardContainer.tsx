import * as React from "react";
import {Dashboard} from "sourcegraph/dashboard/Dashboard";
import {OnboardingContainer} from "sourcegraph/dashboard/OnboardingContainer";

export function DashboardContainer(props: {location: any}): JSX.Element {
	const onboardingStep = props.location.query["ob"] || null;
	return <div>
		{onboardingStep ?
			<OnboardingContainer currentStep={onboardingStep} location={props.location}/> :
			<Dashboard location={props.location}/>}
	</div>;
}
