import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { Integrations } from "sourcegraph/home/Integrations";

interface Props {
	location: RouterLocation;
	router: Router;
}

export function IntegrationsContainer(props: Props): JSX.Element {
	return <LocationStateModal modalName="menuIntegrations" location={props.location} router={props.router}>
		<Integrations location={props.location} />
	</LocationStateModal>;
}
