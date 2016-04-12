import React from "react";
import {Link} from "react-router";

import Component from "sourcegraph/Component";
import DefStore from "sourcegraph/def/DefStore";

import {repoPath, repoRev} from "sourcegraph/repo";
import {urlToTree} from "sourcegraph/tree/routes";
import breadcrumb from "sourcegraph/util/breadcrumb";

import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/breadcrumb.css";

class DefNavContext extends Component {

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repo = repoPath(props.params.splat[0]);
		state.rev = repoRev(props.params.splat[0]);
		state.def = DefStore.defs.get(state.repo, state.rev, props.params.splat[1]);
	}

	render() {
		if (!this.state.def) return null;

		let defFileParts = this.state.def.File.split("/");
		let pathBreadcrumb = breadcrumb(
			`/${this.state.def.File}`,
			(i) => <span key={i} styleName="sep">/</span>,
			(path, component, i, isLast) => (
				!isLast ? <Link to={urlToTree(this.state.repo, this.state.rev, defFileParts.slice(0, i))}
					key={i}
					styleName="inactive">
					{component}
				</Link> :
				<span key={i}>{component}</span>
			)
		);

		return <span>{pathBreadcrumb}</span>;
	}
}

export default CSSModules(DefNavContext, styles);
