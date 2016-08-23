// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Container} from "sourcegraph/Container";
import {DefStore} from "sourcegraph/def/DefStore";
import * as styles from "sourcegraph/def/styles/Def.css";
import {defPath} from "sourcegraph/def/index";
import * as DefActions from "sourcegraph/def/DefActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Button} from "sourcegraph/components/Button";
import {urlToDefInfo} from "sourcegraph/def/routes";

interface Props {
	def: any;
	rev?: string;
	refLocations?: any;
	path?: string;
	location?: any;
}

type State = any;

export class DefPopup extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		features: React.PropTypes.object.isRequired,
	};

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.defObj = props.def;
		state.repo = props.def ? props.def.Repo : null;
		state.rev = props.rev || null;
		state.commitID = props.def ? props.def.CommitID : null;
		state.def = props.def ? defPath(props.def) : null;

		state.authors = DefStore.authors.get(state.repo, state.commitID, state.def);
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.def !== nextState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.commitID, nextState.def));
		}
	}

	stores(): FluxUtils.Store<any>[] {
		return [DefStore];
	}

	render(): JSX.Element | null {
		return (
			<div>
				<Link to={urlToDefInfo(this.state.defObj, this.state.rev)}>
					<Button className={styles.view_all_button} color="blue">View all references</Button>
				</Link>
			</div>
		);
	}
}
