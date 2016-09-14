// tslint:disable typedef ordered-imports
import * as React from "react";
import { abs } from "sourcegraph/app/routePatterns";
import { Container } from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { EventListener } from "sourcegraph/Component";
import { DefStore } from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";
import { Store } from "sourcegraph/Store";
import { urlToDefInfo } from "sourcegraph/def/routes";
import { urlToTree } from "sourcegraph/tree/routes";
import { context } from "sourcegraph/app/context";
import { InjectedRouter } from "react-router";

type State = any;

export function desktopContainer(Component) {
	class DesktopContainer extends Container<{}, State> {
		static contextTypes: React.ValidationMap<any> = {
			router: React.PropTypes.object.isRequired,
		};

		context: { router: InjectedRouter };

		constructor(props: {}) {
			super(props);
			this.desktopNavigation = this.desktopNavigation.bind(this);
			this.state = {
				defSpec: {},
			};
		}

		stores(): Store<any>[] {
			return [DefStore];
		}

		reconcileState(state: State, props: {}): void {
			Object.assign(state, props);
		}

		onStateTransition(oldState: State, newState: State): void {
			const defSpec = newState.defSpec;
			const def = DefStore.defs.get(defSpec.repo, null, defSpec.def);
			if (!def) { return; }
			if (def.Error) {
				messageDesktop(def.Error);
			} else {
				window.location.href = urlToDefInfo(def);
			}
			newState.defSpec = {};
		}

		desktopNavigation(event) {
			const info = event.detail;
			if (info.Kind === "package") {
				const url = urlToTree(info.repo, null, info.treePkg);
				(this.context as any).router.push(url);
				return;
			}
			info.def = infoToDef(info);
			this.setState({ defSpec: { repo: info.repo, def: info.def } });
			Dispatcher.Backends.dispatch(new DefActions.WantDef(info.repo, null, info.def));
		}

		render(): JSX.Element {
			if (!context.user && !allowUnauthed(location.pathname)) {
				location.pathname = abs.login;
			}
			return (
				<div>
					<Component {...this.props} />
					<EventListener target={global.document} event="sourcegraph:desktop:navToSym" callback={this.desktopNavigation} />
				</div>
			);
		}
	}

	return DesktopContainer;
}

const unauthedRoutes = new Set([
	abs.login,
	abs.signup,
	abs.forgot,
]);
function allowUnauthed(location: string) {
	location = location[0] === "/" ?
		location.substring(1) :
		location;
	return unauthedRoutes.has(location);
}

function infoToDef(info) {
	return `${info.UnitType}/${info.pkg}/-/${info.sym}`;
}

// This function sends a message to the desktop application. This is obviously
// not ideal, but it is the only practical way to send a message from the
// webview to the desktop app AFAICT.
function messageDesktop(message) {
	const json = JSON.stringify(message);
	// tslint:disable: no-console
	console.debug(json);
}
