import * as autobind from "autobind-decorator";
import * as debounce from "lodash/debounce";
import * as isEqual from "lodash/isEqual";
import * as React from "react";

import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouteProps, Router } from "sourcegraph/app/router";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { AbsoluteLocation } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { registerEditorCallbacks, registerQuickopenListeners, syncEditorWithRouterProps, toggleQuickopen as quickopen, updateEditorArea, updateWorkspace } from "sourcegraph/editor/config";
import { urlWithRev } from "sourcegraph/repo/routes";
import { init } from "sourcegraph/workbench/main";
import { Services } from "sourcegraph/workbench/services";

interface Props extends AbsoluteLocation, RouteProps {
	rev: string | null;
}

// WorkbenchShell loads the workbench and calls init on it.
@autobind
export class WorkbenchShell extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };
	workbench: Workbench;
	services: ServiceCollection;
	listener: number;
	currWorkspace: IWorkspace;

	domRef(parent: HTMLDivElement): void {
		if (!parent) {
			return;
		}

		const { repo } = this.props;
		const { workbench, services, domElement } = init(URIUtils.createResourceURI(repo), { zapRev: this.props.zapRev, zapRef: this.props.zapRef, commitID: this.props.commitID, branch: this.props.branch });
		registerEditorCallbacks();
		this.workbench = workbench;
		this.services = services;
		this.currWorkspace = (this.services.get(IWorkspaceContextService) as IWorkspaceContextService).getWorkspace();
		updateWorkspace(this.props).then(() => {
			parent.appendChild(domElement);
			updateEditorArea(this.props).then(() => this.layout());
		});
	}

	componentWillMount(): void {
		window.onresize = debounce(this.layout, 50);
		document.body.classList.add("monaco-shell", "vs-dark");
	}

	componentDidMount(): void {
		// Sourcegraph controls the visibility of the embedded vscode modal overlay.
		// This can be implemented by vscode, but without knowing all scenarios in which we
		// want to display an overlay we've left it the Sourcegraph application's responsibility for toggling visibilty.
		const modalOverlay = document.querySelector(".workbench-modal-overlay") as any;
		registerQuickopenListeners(
			() => modalOverlay && (modalOverlay.style.visibility = "visible"),
			() => modalOverlay && (modalOverlay.style.visibility = "hidden"),
		);

		const contextService = Services.get(IWorkspaceContextService);
		contextService.onWorkspaceUpdated(workspace => {
			const revState = workspace.revState;
			const newRepo = workspace.resource.authority + workspace.resource.path !== this.props.repo;
			if (revState && !newRepo) {
				if (revState.zapRev && revState.zapRev !== this.props.zapRev) {
					this.context.router.push(urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.zapRev));
					return;
				}
				if (!revState.zapRev && this.props.zapRev) {
					this.context.router.push(urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.commitID || null));
					return;
				}
			} else if (revState && revState.zapRev) {
				this.context.router.push(urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.zapRev));
			}
		});
	}

	componentWillUpdate(nextProps: Props): void {
		if (!isEqual(nextProps, this.props)) {
			syncEditorWithRouterProps(nextProps);
		}
	}

	componentWillUnmount(): void {
		window.onresize = () => void (0);
	}

	layout(): void {
		if (!this.workbench) {
			return;
		}
		if (window.innerWidth <= 768) {
			// Mobile device, width less than 768px.
			this.workbench.setSideBarHidden(true);
		} else {
			this.workbench.setSideBarHidden(false);
		}
		this.workbench.layout();
	}

	toggleQuickopen(event: KeyboardEvent & { target: Node }): void {
		if (event.target.nodeName === "INPUT" || isNonMonacoTextArea(event.target) || event.metaKey || event.ctrlKey) {
			return;
		}
		const slashKeyCode = 191;
		const escapeKeyCode = 27;
		if (!event.shiftKey && (event.key === "/" || event.key === "Escape" || event.keyCode === slashKeyCode || event.keyCode === escapeKeyCode)) {
			quickopen(event.key === "Escape" || event.keyCode === escapeKeyCode);
			event.preventDefault();
		}
	}

	render(): JSX.Element {
		return <div style={{
			height: "100%",
			display: "flex",
			flex: "1 1 100%",
			flexDirection: "column",
		}} ref={this.domRef}>
			<EventListener target={global.document.body} event="keydown" callback={this.toggleQuickopen} />
		</div>;
	}

}
