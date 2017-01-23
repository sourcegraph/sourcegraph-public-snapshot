import * as React from "react";
import { Code, FlexContainer, Heading, List, Panel } from "sourcegraph/components";
import { colorHelpers, colors, typography, whitespace } from "sourcegraph/components/utils";

export function ColorsComponent(): JSX.Element {

	const colorBlocks = Object.keys(colors).map((label, i) => {
		const rgb = colors[label]();
		const hex = colorHelpers.rgbStringToHex(rgb, true);
		return <Color label={label} rgb={rgb} hex={hex} key={i} />;
	});

	return <div style={{ marginBottom: whitespace[4], marginTop: whitespace[4] }}>
		<Heading level={3} style={{ marginBottom: whitespace[4] }}>Colors</Heading>

		<p>
			There are 8 series of colors in the Sourcegraph palette. Each color has a base color, 3 light shades, and 2 dark shades.
		</p>

		<List style={{ marginBottom: whitespace[2] }}>
			<li><strong>Primaries:</strong> Blue, Purple</li>
			<li><strong>Secondaries:</strong> Orange, Green, Yellow, Red</li>
			<li><strong>Neutrals:</strong> BlueGray, Black, Gray, White</li>
		</List>

		<Heading level={4} style={{ marginTop: whitespace[4] }}>Usage</Heading>
		<p style={{ marginBottom: whitespace[4] }}>
			After importing <Code>colors</Code> from <Code>sourcegraph/components/utils</Code>, you can use it's color function to use it. Each color function takes in an alpha opacity value and outputs an RGBA value string. For example, <Code>colors.blue(0.5)</Code> would output <Code>rgba(0,145,234,1)</Code>.
		</p>

		<FlexContainer justify="between" wrap={true} style={{ paddingBottom: whitespace[4] }}>
			{colorBlocks}
		</FlexContainer>
	</div>;
}

function Color({ label, hex, rgb }: { label: string; hex: string; rgb: string; }): JSX.Element {
	return <Panel hoverLevel="low" style={{
		display: "inline-block",
		flex: "0 0 32%",
		marginBottom: whitespace[3],
	}}>
		<div style={{
			backgroundColor: rgb,
			borderTopLeftRadius: 3,
			borderTopRightRadius: 3,
			height: 100,
			width: "100%",
		}}></div>
		<p style={Object.assign({ padding: whitespace[3] }, typography.size[7])}>
			<strong>Usage:</strong> <Code>colors.{label}(a)</Code><br />
			<strong>RGB:</strong> <Code>{rgb}</Code><br />
			<strong>Hex:</strong> <Code>{hex}</Code><br />
		</p>
	</Panel>;
};
