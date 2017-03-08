import * as React from "react";

import { Router } from "sourcegraph/app/router";
import { LocationProps } from "sourcegraph/app/router";
import { LocationStateModal, dismissModal } from "sourcegraph/components/Modal";
import { BetaInterestForm } from "sourcegraph/home/BetaInterestForm";

interface Props extends LocationProps {
	router: Router;
}

export const BetaSignup = (props: Props): JSX.Element => {
	return <LocationStateModal modalName="menuBeta" title="Join our beta program">
		<BetaInterestForm
			location={props.location}
			loginReturnTo="/beta"
			onSubmit={dismissModal("menuBeta", props.location, props.router)} />
	</LocationStateModal>;
};
