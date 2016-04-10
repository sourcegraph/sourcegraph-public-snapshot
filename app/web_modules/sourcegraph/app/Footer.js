import React from "react";
import Component from "sourcegraph/Component";

import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Footer.css";

class Footer extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return (
			<div styleName="footer">
				<div styleName="left-container">
					<div styleName="left-content">
						<img styleName="logo" src={`${context.assetsRoot}/img/sourcegraph-logo-tagline.svg`} />
						<div styleName="address">
							<a styleName="address-line" href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>
							<span styleName="address-line">121 2nd St, Ste 200</span>
							<span styleName="address-line">San Francisco, CA 94105</span>
						</div>
						<span styleName="version-number">
							{`version ${context.buildVars.Version}`}
						</span>
					</div>
				</div>
				<div styleName="right-container">
					<div styleName="list">
						<div styleName="list-header">Company</div>
						<a styleName="list-element" href="/about/">About</a>
						<a styleName="list-element" href="/careers/">Careers</a>
						<a styleName="list-element" href="/blog/">Blog</a>
						<a styleName="list-element" href="mailto:support@sourcegraph.com">Contact</a>
					</div>
					<div styleName="list">
						<div styleName="list-header">Community</div>
						<a styleName="list-element" href="http://www.meetup.com/Sourcegraph-Hacker-Meetup/">Meetups</a>
						<a styleName="list-element" href="https://twitter.com/srcgraph">Twitter</a>
						<a styleName="list-element" href="https://www.facebook.com/sourcegraph">Facebook</a>
						<a styleName="list-element" href="https://www.youtube.com/channel/UCOy2N25-AHqE43XupT9mwZQ">YouTube</a>
					</div>
					<div styleName="list">
						<div styleName="list-header">Initiatives</div>
						<a styleName="list-element" href="https://srclib.org/" _target="_blank">srclib</a>
						<a styleName="list-element" href="https://fair.io/" _target="_blank">fair.io</a>
					</div>
					<div styleName="list">
						<div styleName="list-header">Legal</div>
						<a styleName="list-element" href="/security/">Security</a>
						<a styleName="list-element" href="/privacy/">Privacy</a>
						<a styleName="list-element" href="/legal/">Terms</a>
					</div>
				</div>
			</div>
		);
	}
}
export default CSSModules(Footer, styles);
