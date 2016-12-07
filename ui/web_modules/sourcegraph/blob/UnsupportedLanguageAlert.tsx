import * as React from "react";
import { Panel } from "sourcegraph/components";
import { Alert } from "sourcegraph/components/symbols";
import { whitespace } from "sourcegraph/components/utils/index";

interface Props {
	ext: string | null;
	style?: React.CSSProperties;
}

export function UnsupportedLanguageAlert({ext, style}: Props): JSX.Element {

	const iconSx = {
		fill: "white",
		marginTop: -2,
		marginRight: whitespace[2],
	};

	const sx = Object.assign({
		display: "inline-block",
		padding: `${whitespace[1]} ${whitespace[2]}`,
	}, style);

	return <Panel color="orange" style={sx}>
		<Alert width={14} style={iconSx} />
		{ext
			? <span>.{ext} files are</span>
			: "This language is"
		} not supported
	</Panel>;
};
