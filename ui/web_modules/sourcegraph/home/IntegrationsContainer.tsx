import * as React from "react";
import { InjectedRouter } from "react-router";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { Integrations } from "sourcegraph/home/Integrations";
import { Location } from "sourcegraph/Location";

interface Props {
	location: Location;
	router: InjectedRouter;
}

export function IntegrationsContainer(props: Props): JSX.Element {
	return <LocationStateModal modalName="menuIntegrations" location={props.location} router={props.router}>
		<Integrations location={props.location} />
	</LocationStateModal>;
}
