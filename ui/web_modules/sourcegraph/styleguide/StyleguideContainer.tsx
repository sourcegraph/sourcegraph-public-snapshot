import * as React from "react";
import {Link} from "react-router";
import {Affix, FlexContainer, Heading, Hero, TabItem, Tabs} from "sourcegraph/components";
import {whitespace} from "sourcegraph/components/utils/index";
import {ComponentsContainer} from "sourcegraph/styleguide/ComponentsContainer";
import * as styles from "sourcegraph/styleguide/styles/StyleguideContainer.css";

export class StyleguideContainer extends React.Component<{}, any> {

	render(): JSX.Element | null {

		const navHeadingSx = {
			marginLeft: whitespace[3],
			paddingLeft: whitespace[1],
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
					<Affix offset={20} style={{flex: "0 0 240px", order: 9999}}>
						<Tabs direction="vertical" color="purple" style={{marginLeft: whitespace[5]}}>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#principles"}}>Principles</Link>
							</TabItem>

							<Heading level={7} style={navHeadingSx}>Brand</Heading>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#brand-voice"}}>Voice</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#brand-logo"}}>Logo and Logotype</Link>
							</TabItem>

							<Heading level={7} style={navHeadingSx}>Layout Components</Heading>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#layout-flexcontainer"}}>FlexContainer</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#layout-affix"}}>Affix</Link>
							</TabItem>

							<Heading level={7} style={navHeadingSx}>UI Components</Heading>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-buttons"}}>Buttons</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-checklists"}}>Checklist Items</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-forms"}}>Forms</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-headings"}}>Headings</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-list"}}>Lists</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-panels"}}>Panels</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-stepper"}}>Stepper</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-symbols"}}>Symbols</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-table"}}>Table</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-tabs"}}>Tabs</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-user"}}>User</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-repository-card"}}>Repository Card</Link>
							</TabItem>
						</Tabs>
					</Affix>
					<ComponentsContainer />
				</FlexContainer>
			</div>
		);
	}
}
