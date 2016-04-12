import React from "react";
import Component from "sourcegraph/Component";

import context from "sourcegraph/app/context";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import styles from "./styles/Footer.css";

class Footer extends Component {
	static propTypes = {
		full: React.PropTypes.bool.isRequired,
	};

	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	// _renderMini renders the mini footer, for app pages.
	_renderMini() {
		return (
			<div styleName="mini">
				<ul styleName="left-mini-box">
					<li>
						<Link to="/" styleName="logo-mark-link">
							<img styleName="logo-mark" title="Sourcegraph" alt="Sourcegraph logo" src={`${context.assetsRoot}/img/sourcegraph-mark.svg`} />
						</Link>
					</li>
					<li><Link styleName="link" to="/">Home</Link></li>
					<li><a styleName="link" href="/blog/">Blog</a></li>
				</ul>
				<ul styleName="right-mini-box">
					<li><a styleName="link" href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a></li>
					<li><a styleName="link" href="/security/">Security</a></li>
					<li><a styleName="link" href="/privacy/">Privacy</a></li>
					<li><a styleName="link" href="/legal/">Terms</a></li>
				</ul>
			</div>
		);
	}

	// _renderFull renders the full footer, for the homepage and other
	// informational pages.
	_renderFull() {
		return (
			<div styleName="full">
				<div styleName="left-container">
					<div styleName="left-content">
						<img styleName="logo" src={`${context.assetsRoot}/img/sourcegraph-logo.svg`} />
						<div styleName="address">
							<span styleName="address-line">121 2nd St, Ste 200</span>
							<span styleName="address-line">San Francisco, CA 94105</span>
							<a styleName="address-line" href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>
						</div>
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

	render() {
		return this.props.full ? this._renderFull() : this._renderMini();
	}
}
export default CSSModules(Footer, styles);
