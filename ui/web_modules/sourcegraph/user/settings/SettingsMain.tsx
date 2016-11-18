import * as React from "react";
import {Heading, Panel} from "sourcegraph/components";
import {colors, layout, whitespace} from "sourcegraph/components/utils";
import {Location} from "sourcegraph/Location";
import {OrgContainer} from "sourcegraph/org/OrgContainer";

interface Props { location: Location; }

export function SettingsMain({location}: Props): JSX.Element  {

	const sx = Object.assign({}, layout.container, {
		marginBottom: whitespace[4],
		marginTop: whitespace[4],
		maxWidth: 960,
		width: "90%",
	});

	return <Panel style={sx} hoverLevel="low" hover={false}>
		<Heading level={5} style={{
			marginTop: whitespace[3],
			marginBottom: whitespace[3],
			marginLeft: whitespace[4],
			marginRight: whitespace[4],
		}}>Organization settings</Heading>
		<hr style={{borderColor: colors.coolGray4(0.7)}} />
		<OrgContainer location={location} />
	</Panel>;
}
