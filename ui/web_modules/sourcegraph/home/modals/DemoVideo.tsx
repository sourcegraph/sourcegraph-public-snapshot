import * as React from "react";

/* TODO(chexee): abstract the presentational component from Modal */
import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";

interface Props {
	location: Object;
}

export const DemoVideo = (props: Props) => {
	const sx = {
		maxWidth: "860px",
		minWidth: "430px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return(
		<LocationStateModal modalName="demo_video" location={props.location}>
			<div className={styles.modal} style={sx}>
				<iframe width="100%" style={{minHeight: "500px"}} src="https://www.youtube.com/embed/tf93F2nc3Yo?rel=0&amp;showinfo=0" frameBorder="0" allowFullScreen={true}></iframe>
			</div>
		</LocationStateModal>
	);
};
