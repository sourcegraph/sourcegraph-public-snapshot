import * as autobind from "autobind-decorator";
import * as debounce from "lodash/debounce";
import * as isEqual from "lodash/isEqual";
import * as React from "react";

import { IDisposable } from "vs/base/common/lifecycle";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IWorkspace, IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouteProps, Router } from "sourcegraph/app/router";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { AbsoluteLocation } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { registerEditorCallbacks, registerQuickopenListeners, syncEditorWithRouterProps, toggleQuickopen as quickopen } from "sourcegraph/editor/config";
import { urlWithRev } from "sourcegraph/repo/routes";
import { init, unmount } from "sourcegraph/workbench/main";
import { Services } from "sourcegraph/workbench/services";

// WorkbenchShell loads the workbench and calls init on it.
@autobind
export class WorkbenchShell extends React.Component<AbsoluteLocation & RouteProps & { rev: string | null }, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };
	workbench: Workbench;
	services: ServiceCollection;
	listener: number;
	disposables: IDisposable[];
	currWorkspace: IWorkspace;

	domRef(domElement: HTMLDivElement): void {
		if (!domElement) {
			if (this.workbench) {
				this.workbench.dispose();
			}
			return;
		}

		const { repo, commitID, path } = this.props;
		const resource = URIUtils.pathInRepo(repo, commitID, path);
		[this.workbench, this.services] = init(domElement, resource, this.props.zapRef);
		registerEditorCallbacks();

		this.layout();
		syncEditorWithRouterProps(this.props);

		this.currWorkspace = (this.services.get(IWorkspaceContextService) as IWorkspaceContextService).getWorkspace();
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
		this.disposables = registerQuickopenListeners(() => modalOverlay.style.visibility = "visible", () => modalOverlay.style.visibility = "hidden");

		const contextService = Services.get(IWorkspaceContextService);
		this.disposables.push(contextService.onWorkspaceUpdated(workspace => {
			const revState = workspace.revState;
			if (revState) {
				if (revState.zapRef && revState.zapRef !== this.props.zapRef) {
					// extension has changed to a new zap ref, update URL
					// this.context.router.push(urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.zapRef));
					window.location.href = urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.zapRef);
					return;
				}
				if (!revState.zapRef && this.props.zapRef) {
					// currently looking at a zap ref, update URL now to the base branch.
					// this.context.router.push(urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.commitID || null));
					window.location.href = urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, revState.commitID || null);
					return;
				}
			}
		}));

		if (this.props.zapRef) {
			// HACK(john): synchronize on zap extension bootstrap
			setTimeout(() => {
				contextService.setWorkspace({
					...contextService.getWorkspace(),
					revState: {
						zapRef: this.props.zapRef,
						commitID: this.props.commitID,
						branch: this.props.branch,
					},
				});
			}, 1000);
		}
	}

	componentWillUnmount(): void {
		window.onresize = () => void (0);
		if (this.disposables) {
			this.disposables.forEach(disposable => disposable.dispose());
		}
		unmount();
	}

	componentWillReceiveProps(nextProps: AbsoluteLocation): void {
		if (!isEqual(nextProps, this.props)) {
			syncEditorWithRouterProps(nextProps);
		}
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

		// HACK: our slightly-larger-than-vscode's status bar needs a re-layout to render
		// entirely within the window, but the layout has to be async. We should update
		// vscode CSS to accomodate a taller status bar so this is unnecessary.
		setTimeout(() => {
			this.workbench.layout();
		}, 100);
	}

	toggleQuickopen(event: KeyboardEvent & { target: Node }): void {
		if (event.target.nodeName === "INPUT" || isNonMonacoTextArea(event.target) || event.metaKey || event.ctrlKey) {
			return;
		}
		const slashKeyCode = 191;
		const escapeKeyCode = 27;
		if (!event.shiftKey && (event.key === "/" || event.key === "Escape" || event.keyCode === slashKeyCode || event.keyCode === escapeKeyCode)) {
			quickopen();
			event.preventDefault();
		}
	}

	render(): JSX.Element {
		this.layout();
		return <div style={{
			height: "100%",
			flex: "1 1 100%",
		}} ref={this.domRef}>
			<EventListener target={global.document.body} event="keydown" callback={this.toggleQuickopen} />
		</div>;
	}

}
