// tslint:disable: typedef ordered-imports curly

import * as React from "react";

// This component renders the Sourcegraph logo, logomark, or logomark with tagline at custom sizes.

type Props = {
	className?: string,
	type?: string,
	width?: string,
};

export class Logo extends React.Component<Props, any> {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		const {width, type, className} = this.props;

		let logoImg = "sourcegraph-mark.svg";
		if (type === "logotype") logoImg = "sourcegraph-logo.svg";
		if (type === "logotype-with-tag") logoImg = "sourcegraph-logo-tagline.svg";

		return <img src={`${(this.context as any).siteConfig.assetsRoot}/img/${logoImg}`} width={width} className={className} title="Sourcegraph" alt="Sourcegraph Logo" />;
	}
}
