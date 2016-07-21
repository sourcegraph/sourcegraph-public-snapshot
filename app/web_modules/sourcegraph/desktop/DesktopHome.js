// @flow

import React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import layout from "sourcegraph/components/styles/_layout.css";
import colors from "sourcegraph/components/styles/_colors.css";
import typography from "sourcegraph/components/styles/_typography.css";
import styles from "./styles/home.css";

import {Heading, List} from "sourcegraph/components";
import {Cone} from "sourcegraph/components/symbols";


class DesktopHome extends React.Component {

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	render() {
		// Switch these out based on detected OS
		const macShortcut = <span><span styleName="label-blue">⌘</span> + <span styleName="label-blue">Shift</span> + <span styleName="label-blue">;</span></span>;
		const windowsShortcut = <span><span styleName="label-blue">CTRL</span> + <span styleName="label-blue">Shift</span> + <span styleName="label-blue">;</span></span>;

		return (
			<div className={`${layout.containerFixed} ${base.pv5} ${base.ph4}`} style={{maxWidth: "560px"}}>
				<Heading align="center" level="4" underline="blue">
					See live examples, search code, and view inline
					<br className={base["hidden-s"]} />&nbsp;documentation to write better code, faster
				</Heading>

				<img src={`${this.context.siteConfig.assetsRoot}/img/sg-desktop.gif`} width="356" title="Usage examples right in your editor" alt="Usage examples right in your editor" style={{maxWidth: "100%", display: "block", imageRendering: "pixelated"}} className={base.center}/>

				<div className={base.mv4}>
					<Heading level="5">Go definitions and usages as you code</Heading>
					<p>
						Install one of our editor integrations, and as you write Go code, this pane will update with contextually relevant information.
					</p>
				</div>
				<div className={base.mv4}>
					<Heading level="5">Semantic, global code search</Heading>
					<p>
						Just hit {windowsShortcut} or click the search box at the top of this page to semantically search for functions and symbols.
					</p>
				</div>
				<div className={base.mv4}>
					<Heading level="5">Powerful search for your private code</Heading>
					<p>
						To enable semantic search and usage examples for your private code, authorize Sourcegraph to access your private repositories.
					</p>
				</div>
				<div className={`${base.mt5} ${typography.f7}`}>
					<Heading level="6">
						<Cone width={16} className={`${colors["fill-orange"]} ${base.mr2}`} style={{
							verticalAlign: "baseline",
							position: "relative",
							top: "1px",
						}} />
						Sourcegraph Desktop is currently in beta
					</Heading>
					<p>
						If the app is not working as expected, see our GitHub to:
					</p>
					<List className={base.mv3}>
						<li><strong><a href="https://github.com/sourcegraph/sourcegraph-desktop#sourcegraph-desktop">Browse troubleshooting tips</a></strong></li>
						<li><strong><a href="https://github.com/sourcegraph/sourcegraph-desktop/issues/new">File an issue</a></strong></li>
					</List>
					<p>
						We love feedback! Shoot us an email at <strong><a href="URL TOEMAIL TEMPLATE">support@sourcegraph.com</a></strong> with ideas on how we can make Sourcegraph Desktop better.
					</p>
				</div>
			</div>
		);
	}
}

export default CSSModules(DesktopHome, styles, {allowMultiple: true});
