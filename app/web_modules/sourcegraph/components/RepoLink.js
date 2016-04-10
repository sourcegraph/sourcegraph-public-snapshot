import React from "react";
import {Link} from "react-router";
import {urlToRepo} from "sourcegraph/repo/routes";
import breadcrumb from "sourcegraph/util/breadcrumb";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/breadcrumb.css";

class RepoLink extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let trimmedPath = this.state.repo;
		if (trimmedPath.indexOf("sourcegraph.com/") !== -1) {
			trimmedPath = trimmedPath.substring("sourcegraph.com/".length);
		}
		if (trimmedPath.indexOf("github.com/") !== -1) {
			trimmedPath = trimmedPath.substring("github.com/".length);
		}
		let pathBreadcrumb = breadcrumb(
			trimmedPath,
			(i) => <span key={i} styleName="sep">/</span>,
			(path, component, i, isLast) => (
				isLast && !this.state.disabledLink ?
					<Link to={urlToRepo(this.state.repo)}
						title={trimmedPath}
						key={i}
						styleName={isLast ? "active" : "inactive"}>
						{component}
					</Link> :
					<span key={i}>{component}</span>
			),
		);

		return <span>{pathBreadcrumb}</span>;
	}
}

RepoLink.propTypes = {
	repo: React.PropTypes.string.isRequired,
	disabledLink: React.PropTypes.bool,
};

export default CSSModules(RepoLink, styles);
