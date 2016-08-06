// tslint:disable

import * as React from "react";
import {Link} from "react-router";
import {urlToRepo, urlToRepoRev} from "sourcegraph/repo/routes";
import breadcrumb from "sourcegraph/util/breadcrumb";
import stripDomain from "sourcegraph/util/stripDomain";

import CSSModules from "react-css-modules";
import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "./styles/breadcrumb.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class RepoLink extends React.Component<any, any> {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		disabledLink: React.PropTypes.bool,
		className: React.PropTypes.string,
	}

	static contextTypes = {
		eventLogger: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		let trimmedPath = stripDomain(this.props.repo);
		let pathBreadcrumb = breadcrumb(
			trimmedPath,
			(i) => <span key={i} styleName="sep" className={base.mh1}>/</span>,
			(path, component, i, isLast) => (
				isLast && !this.props.disabledLink ?
					<Link to={this.props.rev ? urlToRepoRev(this.props.repo, this.props.rev) : urlToRepo(this.props.repo)}
						title={trimmedPath}
						key={i}
						styleName={isLast ? "active" : "inactive"}
						onClick={() => (this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_CLICK, "RepoClicked", {repoName: trimmedPath})}>
						{component}
					</Link> :
					<span key={i}>{component}</span>
			),
		);

		return <span className={this.props.className}>{pathBreadcrumb}</span>;
	}
}

export default CSSModules(RepoLink, styles);
