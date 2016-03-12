import React from "react";

import Component from "sourcegraph/Component";
import RevSwitcherContainer from "sourcegraph/repo/RevSwitcherContainer";
import BuildIndicator from "sourcegraph/build/BuildIndicator";

class BlobToolbar extends Component {
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
		this.state.path.split("/").forEach((seg, i) => {
			basePath += `/${seg}`;
			breadcrumb.push(<span key={i} className="path-component"> / <a href={basePath}>{seg}</a></span>);
		});

		return (
			<div className="code-file-toolbar" ref="toolbar">
				<div className="file-breadcrumb">{breadcrumb}</div>
				<div className="actions">
					<BuildIndicator repo={this.state.repo} commitID={this.state.rev} builds={this.state.builds} />

					<RevSwitcherContainer repo={this.state.repo}
						rev={this.state.rev}
						path={this.state.path}
						route="tree"
						alignRight={true} />
				</div>
			</div>
		);
	}
}

BlobToolbar.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
	path: React.PropTypes.string.isRequired,

	// builds is BuildStore.builds.
	builds: React.PropTypes.object.isRequired,
};

export default BlobToolbar;
