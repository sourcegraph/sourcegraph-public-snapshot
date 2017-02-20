import * as React from "react";
import { Code, FlexContainer, Heading, Panel, Table } from "sourcegraph/components";
import { GitHubLogo, Spinner } from "sourcegraph/components/symbols";
import { typography, whitespace } from "sourcegraph/components/utils/index";

export function Symbols(): JSX.Element {
	return (
		<div style={{ marginBottom: whitespace[4], marginTop: whitespace[4] }}>
			<Heading level={3} style={{ marginBottom: whitespace[2] }}>Symbols</Heading>

			<p>
				We use a subset of the <a href="http://www.parakeet.co/primaries/">Primaries</a> icon set. Each of these icons are individual components that share the same props API. See <Code>/components/symbols</Code> for component usage. Each of these icons can be imported from <Code>sourcegraph/components/symbols/Primaries</Code>.
			</p>

			<Heading level={4} style={{ marginBottom: whitespace[2] }}>Other Symbols</Heading>
			<FlexContainer justify="between" wrap={true} style={{ marginTop: whitespace[4] }}>
				<SymbolTile symbol={GitHubLogo} size={32} name="GitHubLogo" key="GitHubLogo" />
				<SymbolTile symbol={Spinner} size={32} name="Spinner" key="Spinner" />
			</FlexContainer>

			<Heading level={6} style={{ marginTop: whitespace[4], marginBottom: whitespace[3] }}>Properties</Heading>
			<Panel hoverLevel="low" style={{ padding: whitespace[4] }}>
				<Table style={{ width: "100%" }}>
					<thead>
						<tr>
							<td>Prop</td>
							<td>Default value</td>
							<td>Values</td>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td><Code>color</Code></td>
							<td><Code>inherit (black)</Code></td>
							<td>
								any color import from <Code>colors.tsx</Code>
							</td>
						</tr>
						<tr>
							<td><Code>width</Code></td>
							<td><Code>16</Code></td>
							<td>
								any number
							</td>
						</tr>
					</tbody>
				</Table>
			</Panel>
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
