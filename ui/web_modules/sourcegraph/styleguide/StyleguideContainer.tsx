import * as React from "react";
import { Affix, FlexContainer, Heading, Hero, TabItem, Tabs } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils/index";
import { ComponentsContainer } from "sourcegraph/styleguide/ComponentsContainer";
import * as styles from "sourcegraph/styleguide/styles/StyleguideContainer.css";

export function StyleguideContainer(props: {}): JSX.Element {
	const navHeadingSx = {
		marginLeft: whitespace[3],
		paddingLeft: whitespace[3],
		marginTop: whitespace[4],
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
					<Tabs style={{ marginLeft: whitespace[5] }} direction="vertical">
						<TabItem direction="vertical" color="purple">
							<a href="#principles">Principles</a>
						</TabItem>

						<Heading level={7} style={navHeadingSx}>Brand</Heading>
						<TabItem direction="vertical" color="purple">
							<a href="#brand-voice">Voice</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#brand-logo">Logo and Logotype</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#brand-colors">Colors</a>
						</TabItem>

						<Heading level={7} style={navHeadingSx}>Layout Components</Heading>
						<TabItem direction="vertical" color="purple">
							<a href="#layout-flexcontainer">FlexContainer</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#layout-affix">Affix</a>
						</TabItem>

						<Heading level={7} style={navHeadingSx}>UI Components</Heading>
						<TabItem direction="vertical" color="purple">
							<a href="#components-buttons">Buttons</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-forms">Forms</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-headings">Headings</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-list">Lists</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-panels">Panels</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-symbols">Symbols</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-table">Table</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-tabs">Tabs</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-user">User</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-repository-card">Repository Card</a>
						</TabItem>
						<TabItem direction="vertical" color="purple">
							<a href="#components-organization-card">Organization Card</a>
						</TabItem>
					</Tabs>
				</Affix>
				<ComponentsContainer />
			</FlexContainer>
		</div>
	);
}
