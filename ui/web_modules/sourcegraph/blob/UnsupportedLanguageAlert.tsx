import * as React from "react";
import { Panel } from "sourcegraph/components";
import { Alert } from "sourcegraph/components/symbols";
import { whitespace } from "sourcegraph/components/utils/index";

interface Props { ext: string | null; }

export function UnsupportedLanguageAlert({ext}: Props): JSX.Element {

	const iconSx = {
		fill: "white",
		marginTop: -2,
		marginRight: whitespace[2],
	};

	const sx = {
		margin: whitespace[2],
		padding: `${whitespace[2]} ${whitespace[3]}`,
	};

	return <Panel color="orange" style={sx}>
		<Alert width={14} style={iconSx} />
		{ext
			? <span>.{ext} files are</span>
			: "This language is"
		} not supported
	</Panel>;
};
