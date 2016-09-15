// tslint:disable typedef ordered-imports
import * as React from "react";
import { abs } from "sourcegraph/app/routePatterns";
import { Container } from "sourcegraph/Container";
import { EventListener } from "sourcegraph/Component";
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

		reconcileState(state: State, props: {}): void {
			Object.assign(state, props);
		}

		onStateTransition(oldState: State, newState: State): void {
			// TODO(monaco): navigate to def or handle error
		}

		desktopNavigation(event) {
			const info = event.detail;
			if (info.Kind === "package") {
				const url = urlToTree(info.repo, null, info.treePkg);
				(this.context as any).router.push(url);
				return;
			}
			// TODO(monaco): navigate to def
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
