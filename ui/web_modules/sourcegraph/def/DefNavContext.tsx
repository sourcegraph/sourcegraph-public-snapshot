// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";

import {Container} from "sourcegraph/Container";
import {DefStore} from "sourcegraph/def/DefStore";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {urlToTree} from "sourcegraph/tree/routes";
import {breadcrumb} from "sourcegraph/util/breadcrumb";

import * as styles from "sourcegraph/components/styles/breadcrumb.css";

type Props = {
	repo: string,
	rev?: string,
	commitID?: string,
	params: any,
}

export class DefNavContext extends Container<Props, any> {
	reconcileState(state, props: Props) {
		state.repo = props.repo;
		state.rev = props.rev;
		state.commitID = props.commitID;

		state.srclibDataVersion = props.commitID ? TreeStore.srclibDataVersions.get(state.repo, props.commitID) : null;

		const defPath = props.params.splat[1];
		state.defPos = state.srclibDataVersion && state.srclibDataVersion.CommitID ? DefStore.defs.getPos(state.repo, state.srclibDataVersion.CommitID, defPath) : null;
	}

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || (!nextState.srclibDataVersion && prevState.srclibDataVersion !== nextState.srclibDataVersion)) {
			if (nextState.commitID && !nextState.srclibDataVersion) {
				Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.commitID));
			}
		}

		// Rely on the main page's components to get the def, which populates the
		// defPos (if it isn't already populated).
	}

	stores() { return [DefStore, TreeStore]; }

	render(): JSX.Element | null {
		if (!this.state.defPos || this.state.defPos.Error) {
			return null;
		}

		let defFileParts = this.state.defPos.File.split("/");
		let pathBreadcrumb = breadcrumb(
			`/${this.state.defPos.File}`,
			(i) => <span key={i} className={styles.sep}>/</span>,
			(path, component, i, isLast) => (
				!isLast ? <Link to={urlToTree(this.state.repo, this.state.rev, defFileParts.slice(0, i))}
					key={i}
					className={styles.inactive}>
					{component}
				</Link> :
				<span key={i}>{component}</span>
			)
		);

		return <span>{pathBreadcrumb}</span>;
	}
}
