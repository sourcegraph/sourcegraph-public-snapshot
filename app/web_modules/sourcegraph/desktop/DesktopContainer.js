import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import DefStore from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";

import {urlToDefInfo} from "sourcegraph/def/routes";
import {urlToTree} from "sourcegraph/tree/routes";

export default function desktopContainer(Component) {
	class DesktopContainer extends Container {
		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		constructor(props) {
			super(props);
			if (document) {
				document.addEventListener("sourcegraph:desktop:navto", this.desktopNavigation.bind(this));
			}
			this.state = {
				defInfo: {},
			};
		}

		stores() { return [DefStore]; }

		reconcileState(state, props) {
			Object.assign(state, props);
		}

		onStateTransition(state, newState) {
			const info = newState.defInfo;
			const def = DefStore.defs.get(info.Repo, null, info.def);
			if (!def) { return; }
			if (def.Error) {
				messageDesktop(def.Error);
			} else {
				const url = urlToDefInfo(info);
				this.context.router.push(url);
			}
			newState.defInfo = {};
		}

		desktopNavigation({detail}) {
			const info = JSON.parse(detail.definfo);
			if (info.Kind === "package") {
				// TODO detect 404s on bad tree URL.
				const url = urlToTree(info.Repo, null, info.Package);
				this.context.router.push(url);
				return;
			}
			info.def = infoToDef(info);
			this.setState({defInfo: info});
			Dispatcher.Backends.dispatch(new DefActions.WantDef(info.Repo, null, info.def));
		}

		render() {
			return <Component {...this.props}/>;
		}
	}

	return DesktopContainer;
}

function infoToDef(defInfo) {
	return `${defInfo.UnitType}/${defInfo.Unit}/-/${defInfo.Path}`;
}

// This function sends a message to the desktop application.
// Please take care to update the desktop app with any changes
// you make here.
function messageDesktop(message) {
	const json = JSON.stringify(message);
	console.debug(json);
}
