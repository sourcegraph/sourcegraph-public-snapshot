import * as React from "react";
import { Heading, TabItem, Tabs } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils";
import { RepositoryTabs } from "sourcegraph/dashboard";

interface Props {
	active: RepositoryTabs;
	style?: React.CSSProperties;
	setActive: (active: RepositoryTabs) => void;
}

export function TabBar({ active, style, setActive }: Props): JSX.Element {

	const sx = {
		boxSizing: "border-box",
		padding: whitespace[5],
		paddingLeft: 0,
		paddingTop: 0,
		...style,
	};

	const tabSx = { paddingLeft: `calc(${whitespace[4]} - 3px)` };

	return <div style={sx}>
		<Heading level={7} color="white" style={{
			padding: whitespace[4],
			paddingBottom: whitespace[2],
		}}>Repositories</Heading>

		<Tabs>
			<TabItem active={active === "mine"} direction="vertical" inverted={true} style={tabSx}>
				<a onClick={() => setActive("mine")}>Mine</a>
			</TabItem>
			<TabItem active={active === "starred"} direction="vertical" inverted={true} style={tabSx}>
				<a onClick={() => setActive("starred")}>Starred</a>
			</TabItem>
		</Tabs>
	</div>;
}
