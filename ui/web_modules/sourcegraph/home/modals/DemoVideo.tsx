import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";
import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";

interface Props {
	location: Location;
	router: InjectedRouter;
}

export function DemoVideo (props: Props): JSX.Element {
	const sx = {
		maxWidth: "860px",
		minWidth: "430px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return <LocationStateModal modalName="demo_video" location={props.location} router={props.router}>
		<div className={styles.modal} style={sx}>
			<iframe width="100%" style={{minHeight: "500px"}} src="https://www.youtube.com/embed/tf93F2nc3Yo?rel=0&amp;showinfo=0" frameBorder="0" allowFullScreen={true}></iframe>
		</div>
	</LocationStateModal>;
};
