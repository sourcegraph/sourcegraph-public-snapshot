import * as React from "react";
import { Affix, FlexContainer, Heading, Hero, TabItem, Tabs } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils/index";
import { ComponentsContainer } from "sourcegraph/styleguide/ComponentsContainer";
import * as styles from "sourcegraph/styleguide/styles/StyleguideContainer.css";

export function StyleguideContainer(props: {}): JSX.Element {
	const navHeadingSx = {
		marginLeft: whitespace[3],
		paddingLeft: whitespace[3],
		marginTop: whitespace[5],
	};

	return (
		<div className={styles.bg_near_white}>
			<Hero color="purple" pattern="objects">
				<Heading level={2} color="white">The Graph Guide</Heading>
				<p style={{
					marginLeft: "auto",
					marginRight: "auto",
					maxWidth: 560,
					textAlign: "center",
				}}>
					Welcome to the Graph Guide – a living guide to Sourcegraph's brand identity, voice, visual style, and approach to user experience and user interfaces.
				</p>
			</Hero>
			<FlexContainer className={styles.container_fixed}>
				<Affix offset={20} style={{ flex: "0 0 240px", order: 9999 }}>
					<Tabs style={{ marginLeft: whitespace[7] }} direction="vertical">
						<MenuItem anchor="#principles" name="Principles" />

						<Heading level={7} style={navHeadingSx}>Brand</Heading>
						<MenuItem anchor="#brand-voice" name="Voice" />
						<MenuItem anchor="#brand-logo" name="Logo and Logotype" />
						<MenuItem anchor="#brand-colors" name="Colors" />

						<Heading level={7} style={navHeadingSx}>Layout Components</Heading>
						<MenuItem anchor="#layout-flexcontainer" name="FlexContainer" />
						<MenuItem anchor="#layout-affix" name="Affix" />

						<Heading level={7} style={navHeadingSx}>UI Components</Heading>
						<MenuItem anchor="#components-buttons" name="Buttons" />
						<MenuItem anchor="#components-forms" name="Forms" />
						<MenuItem anchor="#components-headings" name="Headings" />
						<MenuItem anchor="#components-list" name="Lists" />
						<MenuItem anchor="#components-panels" name="Panels" />
						<MenuItem anchor="#components-symbols" name="Symbols" />
						<MenuItem anchor="#components-table" name="Table" />
						<MenuItem anchor="#components-tabs" name="Tabs" />
						<MenuItem anchor="#components-user" name="User" />
						<MenuItem anchor="#components-repository-card" name="Repository Card" />
						<MenuItem anchor="#components-organization-card" name="Organization Card" />
					</Tabs>
				</Affix>
				<ComponentsContainer />
			</FlexContainer>
		</div>
	);
}

function MenuItem({ name, anchor }: { anchor: string, name: string }): JSX.Element {
	return <TabItem direction="vertical" color="purple" style={{ paddingLeft: whitespace[5] }}><a href={anchor}>{name}</a></TabItem>;
}
