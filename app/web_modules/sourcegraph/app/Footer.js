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
					<li styleName="item"><a styleName="link" href="https://text.sourcegraph.com">Blog</a></li>
					<li styleName="item"><Link styleName="link" to="/about">About</Link></li>
					<li styleName="item"><Link styleName="link" to="/pricing">Pricing</Link></li>
					<li styleName="item"><a styleName="link" href="https://boards.greenhouse.io/sourcegraph" target="_blank">We're hiring</a></li>
				</ul>
				<ul styleName="right-box">
					<li styleName="item"><Link styleName="link" to="/contact">Contact</Link></li>
					<li styleName="item"><Link styleName="link" to="/security">Security</Link></li>
					<li styleName="item"><Link styleName="link" to="/-/privacy">Privacy</Link></li>
					<li styleName="item"><Link styleName="link" to="/-/terms">Terms</Link></li>
				</ul>
			</div>
		);
	}
}
export default CSSModules(Footer, styles);
