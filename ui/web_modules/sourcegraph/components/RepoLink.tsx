// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {urlToRepo, urlToRepoRev} from "sourcegraph/repo/routes";
import {breadcrumb} from "sourcegraph/util/breadcrumb";
import {stripDomain} from "sourcegraph/util/stripDomain";
import * as classNames from "classnames";

import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/components/styles/breadcrumb.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	repo: string;
	rev?: string;
	disabledLink?: boolean;
	className?: string;
	style?: Object;
}

type State = any;

export class RepoLink extends React.Component<Props, State> {

	render(): JSX.Element | null {
		let trimmedPath = stripDomain(this.props.repo);
		let pathBreadcrumb = breadcrumb(
			trimmedPath,
			(i) => <span key={i} className={classNames(styles.sep, base.mh1)}>/</span>,
			(path, component, i, isLast) => (
					<Link to={this.props.rev ? urlToRepoRev(this.props.repo, this.props.rev) : urlToRepo(this.props.repo)}
						title={trimmedPath}
						key={i}
						className={isLast ? styles.active : styles.inactive}
						onClick={() => EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_CLICK, "RepoClicked", {repoName: trimmedPath})}>
						{component}
					</Link>
			),
		);

		return <span style={this.props.style} className={this.props.className}>{pathBreadcrumb}</span>;
	}
}
