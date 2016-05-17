import React from "react";

import {Link} from "react-router";
import Logo from "sourcegraph/components/Logo";
import CSSModules from "react-css-modules";
import styles from "./styles/Footer.css";

class Footer extends React.Component {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	render() {
		return (
			<div styleName="footer">
				<ul styleName="left-box">
					<li styleName="item">
						<Link to="/" styleName="link">
							<Logo width="16px" styleName="logo-mark" />
						</Link>
					</li>
					<li styleName="item"><a styleName="link" href="/blog/">Blog</a></li>
					<li styleName="item"><a styleName="link" href="/about/">About</a></li>
					<li styleName="item"><a styleName="link" href="/careers/">Careers</a></li>
				</ul>
				<ul styleName="right-box">
					<li styleName="item"><a styleName="link" href="mailto:hi@sourcegraph.com">Contact</a></li>
					<li styleName="item"><a styleName="link" href="/security/">Security</a></li>
					<li styleName="item"><a styleName="link" href="/privacy/">Privacy</a></li>
					<li styleName="item"><a styleName="link" href="/legal/">Terms</a></li>
				</ul>
			</div>
		);
	}
}
export default CSSModules(Footer, styles);
