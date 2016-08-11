// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {urlToRepo, urlToRepoRev} from "sourcegraph/repo/routes";
import {breadcrumb} from "sourcegraph/util/breadcrumb";
import {stripDomain} from "sourcegraph/util/stripDomain";
import * as classNames from "classnames";

import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "./styles/breadcrumb.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	repo: string;
	rev?: string;
	disabledLink?: boolean;
	className?: string;
}

export class RepoLink extends React.Component<Props, any> {
	static contextTypes = {
		eventLogger: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		let trimmedPath = stripDomain(this.props.repo);
		let pathBreadcrumb = breadcrumb(
			trimmedPath,
			(i) => <span key={i} className={classNames(styles.sep, base.mh1)}>/</span>,
			(path, component, i, isLast) => (
				isLast && !this.props.disabledLink ?
					<Link to={this.props.rev ? urlToRepoRev(this.props.repo, this.props.rev) : urlToRepo(this.props.repo)}
						title={trimmedPath}
						key={i}
						className={isLast ? styles.active : styles.inactive}
						onClick={() => (this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_CLICK, "RepoClicked", {repoName: trimmedPath})}>
						{component}
					</Link> :
					<span key={i}>{component}</span>
			),
		);

		return <span className={this.props.className}>{pathBreadcrumb}</span>;
	}
}
