// tslint:disable

import * as React from "react";

import RepoLink from "sourcegraph/components/RepoLink";
import {Label} from "sourcegraph/components/index";
import RevSwitcherContainer from "sourcegraph/repo/RevSwitcherContainer";
import * as styles from "./styles/Repo.css";

type Props = {
	repo: string,
	rev?: string,
	commitID?: string,
	inventory?: any,
	repoNavContext: JSX.Element,
	repoObj?: any,
	isCloning: boolean,

	// to pass to RevSwitcherContainer so it can construct URLs
	routes: any[],
	routeParams: any,
};

class NavContext extends React.Component<Props, any> {
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

export default NavContext;
