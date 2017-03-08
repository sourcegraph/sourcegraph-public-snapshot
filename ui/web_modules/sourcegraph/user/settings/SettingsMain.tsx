import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { Heading, Panel } from "sourcegraph/components";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { OrgContainer } from "sourcegraph/org/OrgContainer";

interface Props { location: RouterLocation; }

export function SettingsMain({ location }: Props): JSX.Element {

	const sx = Object.assign({}, layout.container, {
		marginBottom: whitespace[5],
		marginTop: whitespace[5],
		maxWidth: 960,
		width: "90%",
	});

	return <Panel style={sx} hoverLevel="low" hover={false}>
		<Heading level={5} style={{
			marginTop: whitespace[3],
			marginBottom: whitespace[3],
			marginLeft: whitespace[5],
			marginRight: whitespace[5],
		}}>Organization settings</Heading>
		<hr style={{ borderColor: colors.blueGrayL3(0.7) }} />
		<OrgContainer location={location} />
	</Panel>;
}
