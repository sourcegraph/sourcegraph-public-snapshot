import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";

import {CloseIcon} from "sourcegraph/components/Icons";
import {Integrations as IntegrationsContent} from "sourcegraph/home/Integrations";

import * as base from "sourcegraph/components/styles/_base.css";

/* TODO(chexee): abstract the presentational component from Modal */
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";

interface Props {
	location: Location;
	router: InjectedRouter;
}

export const Integrations = (props: Props): JSX.Element => {
	const sx = {
		maxWidth: "440px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return (
		<LocationStateModal modalName="menuIntegrations" location={props.location} style={sx}>
			<div className={styles.modal} style={sx}>
				<a className={styles.modal_dismiss} onClick={dismissModal("menuIntegrations", props.location, props.router)}>
					<CloseIcon className={base.pt2} />
				</a>
				<IntegrationsContent location={props.location}/>
			</div>
		</LocationStateModal>
	);
};
