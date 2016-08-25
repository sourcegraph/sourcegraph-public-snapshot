// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Hero, Heading, FlexContainer, Tabs, TabItem, Affix} from "sourcegraph/components/index";
import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/styleguide/styles/StyleguideContainer.css";
import {ComponentsContainer} from "sourcegraph/styleguide/ComponentsContainer";

export class StyleguideContainer extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div className={styles.bg_near_white}>
				<Hero color="purple" pattern="objects">
					<Heading level="2" color="white">The Graph Guide</Heading>
					<p style={{maxWidth: "560px"}} className={base.center}>
						Welcome to the Graph Guide – a living guide to Sourcegraph's brand identity, voice, visual style, and approach to user experience and user interfaces.
					</p>
				</Hero>
				<FlexContainer className={styles.container_fixed}>
					<Affix offset={20} style={{flex: "0 0 240px"}} className={base.orderlast}>
						<Tabs direction="vertical" color="purple" className={base.ml5}>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#principles"}}>Principles</Link>
							</TabItem>

							<Heading level="6" className={base.mt4}>Brand</Heading>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#brand-voice"}}>Voice</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#brand-logo"}}>Logo and Logotype</Link>
							</TabItem>
							{/* <TabItem>Colors</TabItem>
							<TabItem>Typography</TabItem>}

							{/*  <Heading level="6" className={base.mt4}>Utilities</Heading>
							<TabItem>Padding</TabItem>
							<TabItem>Margin</TabItem>
							<TabItem>Colors</TabItem>
							<TabItem>Layout</TabItem>*/}

							<Heading level="6" className={base.mt4}>Layout Components</Heading>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#layout-flexcontainer"}}>FlexContainer</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#layout-affix"}}>Affix</Link>
							</TabItem>

							<Heading level="6" className={base.mt4}>UI Components</Heading>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-headings"}}>Headings</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-forms"}}>Forms</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-buttons"}}>Buttons</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "components-tabs"}}>Tabs</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-panels"}}>Panels</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-stepper"}}>Stepper</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-checklists"}}>Checklist Items</Link>
							</TabItem>
							<TabItem>
								<Link to={{pathname: "styleguide", search: "#components-table"}}>Table</Link>
							</TabItem>
						</Tabs>
					</Affix>
					<ComponentsContainer />
				</FlexContainer>
			</div>
		);
	}
}
