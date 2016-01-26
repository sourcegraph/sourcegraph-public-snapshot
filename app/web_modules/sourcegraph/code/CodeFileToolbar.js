import React from "react";

import Component from "sourcegraph/Component";
import RepoRevSwitcher from "../../../script/components/RepoRevSwitcher"; // FIXME
import RepoBuildIndicator from "../../../script/components/RepoBuildIndicator"; // FIXME

class CodeFileToolbar extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		// TODO replace with proper shared component
		let revPart = this.state.rev ? `@${this.state.rev}` : "";
		let basePath = `/${this.state.repo}${revPart}/.tree`;
		let repoSegs = this.state.repo.split("/");
		let breadcrumb = [
			<span key="base" className="path-component">
				<a className="path-component" href={basePath}>{repoSegs[repoSegs.length-1]}</a>
			</span>,
		];
		this.state.tree.split("/").forEach((seg, i) => {
			basePath += `/${seg}`;
			breadcrumb.push(<span key={i} className="path-component"> / <a href={basePath}>{seg}</a></span>);
		});

		return (
			<div className="code-file-toolbar container" ref="toolbar">
				<div className="file-breadcrumb">
					<i className={this.state.file ? "fa fa-file" : "fa fa-spinner fa-spin"} />{breadcrumb}
				</div>
				<div className="actions">
					<RepoBuildIndicator btnSize="btn-sm" RepoURI={this.state.repo} commitID={this.state.rev} />

					<RepoRevSwitcher repoSpec={this.state.repo}
						rev={this.state.rev}
						path={this.state.tree}
						alignRight={true} />
				</div>
			</div>
		);
	}
}

export default CodeFileToolbar;
