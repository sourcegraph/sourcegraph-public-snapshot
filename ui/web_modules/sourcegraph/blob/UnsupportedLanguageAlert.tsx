import * as React from "react";
import {Base, Panel} from "sourcegraph/components";
import {Alert} from "sourcegraph/components/symbols";
import {whitespace} from "sourcegraph/components/utils/index";

interface Props { ext: string | null; }

export function UnsupportedLanguageAlert({ext}: Props): JSX.Element {

	const iconSx = {
		fill: "white",
		marginTop: -2,
		marginRight: whitespace[2],
	};

	return <Panel color="orange" inverse={true} style={{margin: whitespace[2]}}>
		<Base px={3} py={1}>
			<Alert width={14} style={iconSx} />
			{ext
				? <span>.{ext} files are</span>
				: "This language is"
			} not fully supported
		</Base>
	</Panel>;
};
