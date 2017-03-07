import * as React from "react";
import { Code, FlexContainer, Heading, Panel } from "sourcegraph/components";
import { GitHubLogo, Spinner } from "sourcegraph/components/symbols";
import { typography, whitespace } from "sourcegraph/components/utils/index";

export function Symbols(): JSX.Element {
	return (
		<div>
			<p>
				We use a subset of the <a href="http://www.parakeet.co/primaries/">Primaries</a> icon set. Each of these icons are individual components that share the same props API. See <Code>/components/symbols</Code> for component usage. Each of these icons can be imported from <Code>sourcegraph/components/symbols/Primaries</Code>.
			</p>

			<Heading level={4} style={{ marginBottom: whitespace[2] }}>Other Symbols</Heading>
			<FlexContainer justify="between" wrap={true} style={{ marginTop: whitespace[4] }}>
				<SymbolTile symbol={GitHubLogo} size={32} name="GitHubLogo" key="GitHubLogo" />
				<SymbolTile symbol={Spinner} size={32} name="Spinner" key="Spinner" />
			</FlexContainer>
		</div>
	);
}

function SymbolTile({ symbol, size, name }: { symbol: any; size: number; name: string; }): JSX.Element {
	return <Panel hoverLevel="low" style={{
		display: "inline-block",
		flex: "0 0 23%",
		marginBottom: whitespace[3],
		padding: whitespace[2],
		textAlign: "center",
	}}>
		{symbol({ width: size })}<br />
		<span style={typography.size[7]}>{name}</span>
	</Panel>;
}
