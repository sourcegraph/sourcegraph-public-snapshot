// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {DefKey} from "sourcegraph/api";
import {DefStore} from "sourcegraph/def/DefStore";
import * as styles from "sourcegraph/def/styles/Def.css";
import {Button} from "sourcegraph/components/Button";
import {urlToDefKeyInfo} from "sourcegraph/def/routes";
import {BlobPos} from "sourcegraph/def/DefActions";
import {Store} from "sourcegraph/Store";

interface Props {
	pos: BlobPos;
}

interface State extends Props {
	defKey: DefKey | null;
};

export class BlobMargin extends Container<Props, State> {
	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		let info: any = DefStore.hoverInfos.get(props.pos);
		if (info && info.def && info.def.CommitID) {
			state.defKey = info.def;
		} else {
			state.defKey = null;
		}
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (nextState.pos && (!prevState.pos || !blobPosEqual(prevState.pos, nextState.pos))) {
			Dispatcher.Backends.dispatch(new DefActions.WantHoverInfo(nextState.pos));
		}
	}

	stores(): Store<any>[] {
		return [DefStore];
	}

	render(): JSX.Element | null {
		if (this.state.defKey == null) {
			return null;
		}
		return (
			<div>
				<Link to={urlToDefKeyInfo(this.state.defKey)}>
					<Button className={styles.view_all_button} color="blue">View all references</Button>
				</Link>
			</div>
		);
	}
}

function blobPosEqual(a: BlobPos, b: BlobPos): boolean {
	return a.repo === b.repo && a.commit === b.commit && a.file === b.file && a.line === b.line && a.character === b.character;
}
