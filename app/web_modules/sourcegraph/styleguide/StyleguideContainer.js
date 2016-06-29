// @flow

import React from "react";
import {Hero, Heading, FlexContainer, Tabs, TabItem, Affix} from "sourcegraph/components";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import styles from "./styles/StyleguideContainer.css";
import ComponentsContainer from "./ComponentsContainer";

class StyleguideContainer extends React.Component {

	_pageLinkScroll(e) {
		const anchorLabel = e.target.getAttribute("data-anchor");
		const anchorEl = document.getElementById(anchorLabel);
		anchorEl.scrollIntoView();
		return false;
	}

	render() {
		return (
			<div styleName="bg-near-white">
				<Hero color="purple" pattern="objects">
					<Heading level="2" color="white">The Graph Guide</Heading>
					<p style={{maxWidth: "560px"}} className={base.center}>
						Welcome to the Graph Guide – a living guide to Sourcegraph's brand identity, voice, visual style, and approach to user experience and user interfaces.
					</p>
				</Hero>
				<FlexContainer styleName="container-fixed">
					<Affix offset={20} style={{flex: "0 0 220px"}} className={base.orderlast}>
						<Tabs direction="vertical" color="purple" className={base.ml5}>
							<TabItem>
								<a data-anchor="principles" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Principles</a>
							</TabItem>

							<Heading level="5" className={base.mt4}>Brand</Heading>
							<TabItem>
								<a data-anchor="brand-voice" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Voice</a>
							</TabItem>
							<TabItem>
								<a data-anchor="brand-logo" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Logo and Logotype</a>
							</TabItem>
							{/* <TabItem>Colors</TabItem>
							<TabItem>Typography</TabItem>}

							{/*  <Heading level="5" className={base.mt4}>Utilities</Heading>
							<TabItem>Padding</TabItem>
							<TabItem>Margin</TabItem>
							<TabItem>Colors</TabItem>
							<TabItem>Layout</TabItem>*/}

							<Heading level="5" className={base.mt4}>Layout Components</Heading>
							<TabItem>
								<a data-anchor="layout-flexcontainer" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>FlexContainer</a>
							</TabItem>
							<TabItem>
								<a data-anchor="layout-affix" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Affix</a>
							</TabItem>

							<Heading level="5" className={base.mt4}>UI Components</Heading>
							<TabItem>
								<a data-anchor="components-headings" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Headings</a>
							</TabItem>
							<TabItem>
								<a data-anchor="components-buttons" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Buttons</a>
							</TabItem>
							<TabItem>
								<a data-anchor="components-tabs" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Tabs</a>
							</TabItem>
							<TabItem>
								<a data-anchor="components-panels" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Panels</a>
							</TabItem>
							<TabItem>
								<a data-anchor="components-stepper" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Stepper</a>
							</TabItem>
							<TabItem>
								<a data-anchor="components-checklists" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Checklist Items</a>
							</TabItem>
							<TabItem>
								<a data-anchor="components-table" style={{cursor: "pointer"}} onClick={(e) => this._pageLinkScroll(e)}>Table</a>
							</TabItem>
						</Tabs>
					</Affix>
					<ComponentsContainer />
				</FlexContainer>
			</div>
		);
	}
}

export default CSSModules(StyleguideContainer, styles, {allowMultiple: true});
