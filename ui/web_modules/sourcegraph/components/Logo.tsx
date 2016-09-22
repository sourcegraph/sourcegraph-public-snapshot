import * as React from "react";
import {context} from "sourcegraph/app/context";

// This component renders the Sourcegraph logo, logomark, or logomark with tagline at custom sizes.

interface Props {
	className?: string;
	type?: string;
	width?: string;
}

type State = any;

export class Logo extends React.Component<Props, State> {

	render(): JSX.Element | null {
		const {width, type, className} = this.props;

		let logoImg = "sourcegraph-mark.svg";
		if (type === "logotype") {
			logoImg = "sourcegraph-logo.svg";
		}
		if (type === "logotype-with-tag") {
			logoImg = "sourcegraph-logo-tagline.svg";
		}

		return <img src={`${context.assetsRoot}/img/${logoImg}`} width={width} className={className} title="Sourcegraph" alt="Sourcegraph Logo" />;
	}
}
