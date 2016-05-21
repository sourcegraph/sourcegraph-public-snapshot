import React from "react";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import TreeStore from "sourcegraph/tree/TreeStore";

import {urlToTree} from "sourcegraph/tree/routes";
import breadcrumb from "sourcegraph/util/breadcrumb";

import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/breadcrumb.css";

class DefNavContext extends Container {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string.isRequired,
		params: React.PropTypes.object.isRequired,
	}

	reconcileState(state, props) {
		state.repo = props.repo;
		state.rev = props.rev;

		const srclibDataVersion = props.commitID ? TreeStore.srclibDataVersions.get(state.repo, props.commitID) : null;
		const defPath = props.params.splat[1];
		state.defPos = srclibDataVersion && srclibDataVersion.CommitID ? DefStore.defs.getPos(state.repo, srclibDataVersion.CommitID, defPath) : null;
	}

	stores() { return [DefStore, TreeStore]; }

	render() {
		if (!this.state.defPos || this.state.defPos.Error) return null;

		let defFileParts = this.state.defPos.File.split("/");
		let pathBreadcrumb = breadcrumb(
			`/${this.state.defPos.File}`,
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
