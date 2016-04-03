// @flow

import React from "react";
import {Link} from "react-router";
import Style from "./styles/Blob.css";
import urlTo from "sourcegraph/util/urlTo";
import breadcrumb from "sourcegraph/util/breadcrumb";
import breadcrumbStyle from "sourcegraph/components/styles/breadcrumb.css";

export default class RepoNavContext extends React.Component {
	static propTypes = {
		params: React.PropTypes.object.isRequired,
	};

	render() {
		let pathBreadcrumb = breadcrumb(
			this.props.params.splat[1],
			(i) => <span key={i} className={breadcrumbStyle.sep}>/</span>,
			(path, component, i, isLast) => (
				<Link to={urlTo("tree", {...this.props.params, splat: path})}
					key={i}
					className={isLast ? Style.activePathComponent : Style.inactivePathComponent}>
					{component}
				</Link>
			),
		);

		return (
			<div className={Style.repoNavContext}>
				{pathBreadcrumb}
			</div>
		);
	}
}
