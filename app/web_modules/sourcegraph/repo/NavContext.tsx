// tslint:disable

import * as React from "react";

import RepoLink from "sourcegraph/components/RepoLink";
import {Label} from "sourcegraph/components/index";
import RevSwitcherContainer from "sourcegraph/repo/RevSwitcherContainer";
import CSSModules from "react-css-modules";
import * as styles from "./styles/Repo.css";

class NavContext extends React.Component<any, any> {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		inventory: React.PropTypes.object,
		repoNavContext: React.PropTypes.element,
		repoObj: React.PropTypes.object,
		isCloning: React.PropTypes.bool.isRequired,

		// to pass to RevSwitcherContainer so it can construct URLs
		routes: React.PropTypes.array.isRequired,
		routeParams: React.PropTypes.object.isRequired,
	};

	_isLanguageUnsupported(): boolean {
		if (!this.props.inventory || !this.props.inventory.Languages || !this.props.inventory.PrimaryProgrammingLanguage) return false; // innocent until proven guilty
		return this.props.inventory.Languages.filter((lang) => ["Go", "Java", "JavaScript", "Shell"].includes(lang.Name)).length === 0;
	}

	render(): JSX.Element | null {
		return (
			<div className={styles.nav}>
				{this._isLanguageUnsupported() &&
					<Label style={{marginRight: "10px"}} color="warning">
						{`${this.props.inventory.PrimaryProgrammingLanguage} is not yet supported`}
					</Label>
				}
				{this.props.repoObj &&
					<RepoLink repo={this.props.repo} rev={this.props.rev} />
				}
				<div className={styles.repo_nav_context}>{this.props.repoNavContext}</div>
				{this.props.commitID && <RevSwitcherContainer
					repo={this.props.repo}
					repoObj={this.props.repoObj}
					rev={this.props.rev}
					commitID={this.props.commitID}
					routes={this.props.routes}
					routeParams={this.props.routeParams}
					isCloning={this.props.isCloning} />}
			</div>
		);
	}
}

export default CSSModules(NavContext, styles);
