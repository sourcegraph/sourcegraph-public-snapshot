import React from "react";
import {Link} from "react-router";
import {urlToRepoRev} from "sourcegraph/repo/routes";
import breadcrumb from "sourcegraph/util/breadcrumb";

import CSSModules from "react-css-modules";
import styles from "./styles/breadcrumb.css";

class RepoLink extends React.Component {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string.isRequired,
		disabledLink: React.PropTypes.bool,
		className: React.PropTypes.string,
	}

	render() {
		let trimmedPath = this.props.repo;
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
				isLast && !this.props.disabledLink ?
					<Link to={urlToRepoRev(this.props.repo, this.props.rev)}
						title={trimmedPath}
						key={i}
						styleName={isLast ? "active" : "inactive"}>
						{component}
					</Link> :
					<span key={i}>{component}</span>
			),
		);

		return <span className={this.props.className}>{pathBreadcrumb}</span>;
	}
}

export default CSSModules(RepoLink, styles);
