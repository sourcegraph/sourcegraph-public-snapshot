// @flow

import React from "react";
import {Hero, Heading} from "sourcegraph/components";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import ComponentsContainer from "./ComponentsContainer";

class StyleguideContainer extends React.Component {

	render() {
		return (
			<div>
				<Hero color="dark" pattern="objects">
					<Heading level="1" color="white" underline="white">The Graph Guide</Heading>
					<p style={{maxWidth: "560px"}} className={base.center}>
						Welcome to the Graph Guide – a living guide to Sourcegraph's brand identity, voice, visual style, and approach to user experience and user interfaces. Everything else TBD – for now, just components.
					</p>
				</Hero>
				<ComponentsContainer />
			</div>
		);
	}
}

export default CSSModules(StyleguideContainer, base);

