import * as classNames from "classnames";
import * as React from "react";
import {Link} from "react-router";

import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/components/styles/breadcrumb.css";
import {urlToRepo, urlToRepoRev} from "sourcegraph/repo/routes";
import {breadcrumb} from "sourcegraph/util/breadcrumb";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {stripDomain} from "sourcegraph/util/stripDomain";

interface Props {
	repo: string;
	rev?: string;
	disabledLink?: boolean;
	className?: string;
	style?: Object;
}

export class RepoLink extends React.Component<Props, {}> {

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
						onClick={() => AnalyticsConstants.Events.Repository_Clicked.logEvent({repoName: trimmedPath})}>
						{component}
					</Link>
			),
		);

		return <span style={this.props.style} className={this.props.className}>{pathBreadcrumb}</span>;
	}
}
