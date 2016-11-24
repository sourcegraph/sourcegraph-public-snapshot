import * as React from "react";
import { context } from "sourcegraph/app/context";

// This component renders the Sourcegraph logo, logomark, or logomark with tagline at custom sizes.

interface Props {
	className?: string;
	type?: "logotype-with-tag" | "logotype";
	width?: string;
	style?: React.CSSProperties;
}

export function Logo({width, type, className, style}: Props): JSX.Element {
	let logoImg = "sourcegraph-mark.svg";
	if (type === "logotype") {
		logoImg = "sourcegraph-logo.svg";
	}
	if (type === "logotype-with-tag") {
		logoImg = "sourcegraph-logo-tagline.svg";
	}

	return <img src={`${context.assetsRoot}/img/${logoImg}`} width={width} className={className} title="Sourcegraph" alt="Sourcegraph Logo" style={style} />;
}
