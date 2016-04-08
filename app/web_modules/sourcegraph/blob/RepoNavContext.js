// @flow

import React from "react";

import Component from "sourcegraph/Component";

import {Link} from "sourcegraph/components";
import urlTo from "sourcegraph/util/urlTo";
import breadcrumb from "sourcegraph/util/breadcrumb";

import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/breadcrumb.css";

class RepoNavContext extends Component {
	static propTypes = {
		params: React.PropTypes.object.isRequired,
	};

	reconcileState(state, props) {
		Object.assign(state, props);
		state.blobPath = props.params.splat[1];
		state.pathParts = state.blobPath.split("/");
	}

	render() {
		let pathBreadcrumb = breadcrumb(
			`/${this.state.blobPath}`,
			(i) => <span key={i} styleName="sep">/</span>,
			(path, component, i, isLast) => (
				<Link to={isLast ?
					urlTo("blob", {...this.state.params}) :
					urlTo("tree", {
						...this.state.params,
						splat: [this.state.params.splat[0], this.state.pathParts.slice(0, i).join("/")],
					})}
					key={i}
					styleName={isLast ? "active" : "inactive"}>
					{component}
				</Link>
			),
		);

		return (
			<span>{pathBreadcrumb}</span>
		);
	}
}

export default CSSModules(RepoNavContext, styles);
