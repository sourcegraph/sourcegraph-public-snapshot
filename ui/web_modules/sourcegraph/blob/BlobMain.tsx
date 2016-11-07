import * as debounce from "lodash/debounce";
import * as React from "react";
import Helmet from "react-helmet";
import { InjectedRouter } from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import "sourcegraph/blob/BlobBackend";
import { BlobStore } from "sourcegraph/blob/BlobStore";
import { BlobTitle } from "sourcegraph/blob/BlobTitle";
import { urlToBlob } from "sourcegraph/blob/routes";
import {FlexContainer, TourOverlay} from "sourcegraph/components";
import {ChromeExtensionToast} from "sourcegraph/components/ChromeExtensionToast";
import {OnboardingModals} from "sourcegraph/components/OnboardingModals";
import {colors} from "sourcegraph/components/utils/colors";
import {Container} from "sourcegraph/Container";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import {URIUtils} from "sourcegraph/core/uri";
import { Location } from "sourcegraph/Location";
import { trimRepo } from "sourcegraph/repo";
import {Store} from "sourcegraph/Store";

// Don't load too much from vscode, because this file is loaded in the
// initial bundle and we want to keep the initial bundle small.
import {Editor} from "sourcegraph/editor/Editor";
import {EditorComponent} from "sourcegraph/editor/EditorComponent";
import {IEditorOpenedEvent} from "sourcegraph/editor/EditorService";
import {IRange} from "vs/editor/common/editorCommon";

interface Props {
	repo: string;
	repoObj: any;
	rev: string;
	commitID: string;
	isCloning: boolean;
	path: string;
	routes: Object[];
	routeParams: RouteParams;
	startLine?: number;
	startCol?: number;
	endLine?: number;
	endCol?: number;
	location: Location;

	// TODO(sqs): now that BlobMain no longer uses the blob directly
	// (the EditorService fetches it), we can save on a network RTT by
	// eliminating the WantFile dispatch.
}

interface State extends Props {
	toast: string | null;
}

// BlobMain wraps the Editor component for the primary code view.
export class BlobMain extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: {
		router: InjectedRouter,
	};

	private _editor?: Editor;
	private _shortCircuitURLNavigationOnEditorOpened: number = 0;

	constructor(props: Props) {
		super(props);
		this._setEditor = this._setEditor.bind(this);
		this._onKeyDownForFindInPage = this._onKeyDownForFindInPage.bind(this);
		this._onResize = debounce(this._onResize.bind(this), 300, { leading: true, trailing: true });
	}

	componentDidMount(): void {
		super.componentDidMount();

		global.document.addEventListener("keydown", this._onKeyDownForFindInPage);
		global.document.addEventListener("resize", this._onResize);

		global.document.body.style.overflowY = "hidden";
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();

		global.document.removeEventListener("keydown", this._onKeyDownForFindInPage);
		global.document.removeEventListener("resize", this._onResize);

		global.document.body.style.overflowY = "auto";
	}

	componentWillReceiveProps(nextProps: Props): void {
		super.componentWillReceiveProps(nextProps, null);

		if (this._editor) {
			this._editorPropsChanged(this.props, nextProps);
		}
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.toast = BlobStore.toast;
	}

	stores(): Store<any>[] {
		return [BlobStore];
	}

	_setEditor(editor: Editor | null): void {
		this._editor = editor || undefined;
		if (this._editor) {
			this._editorPropsChanged(null, this.props);
			this._editor.onDidOpenEditor(e => this._onEditorOpened(e));
			this._editor.onLineSelected((mouseDownEvent, mouseUpEvent): void => this._onLineSelected());
		}
	}

	_editorPropsChanged(prevProps: Props | null, nextProps: Props): void {
		if (!this._editor) {
			throw new Error("editor is not ready");
		}

		if (!prevProps || (prevProps.repo !== nextProps.repo || prevProps.rev !== nextProps.rev || prevProps.commitID !== nextProps.commitID || prevProps.path !== nextProps.path || prevProps.startLine !== nextProps.startLine || prevProps.startCol !== nextProps.startCol || prevProps.endLine !== nextProps.endLine || prevProps.endCol !== nextProps.endCol)) {
			if (nextProps.commitID) {
				// Use absolute commit IDs for the editor model URI.
				// Normalizing repo URI (for example github.com/HvyIndustries/Crane => github.com/HvyIndustries/crane)
				const uri = URIUtils.pathInRepo(nextProps.repoObj && nextProps.repoObj.URI ?  nextProps.repoObj.URI : nextProps.repo, nextProps.commitID, nextProps.path);

				let range: IRange;
				if (typeof nextProps.startLine === "number") {
					const rop = RangeOrPosition.fromOneIndexed(nextProps.startLine, nextProps.startCol, nextProps.endLine, nextProps.endCol);
					range = rop.toMonacoRangeAllowEmpty();
				} else {
					// Without a start line, set the cursor to start at the beginning of the document.
					//
					// By default, set the selction of the cursor to be the end of the
					// first line. Without this, the Ctrl+F search functionality will only
					// search within the selected line, which is the first line by default.
					// There's currently no API for unselecting a line so this is a workaround.
					range = {startLineNumber: 1, startColumn: Infinity, endLineNumber: 1, endColumn: Infinity} as IRange;
				}

				// If you are new to this code, you may be confused how this method interacts with _onEditorOpened.
				// You may also wonder how _shortCircuitURLNavigationOnEditorOpened works. Here's an explanation:
				//
				// Calling this._editor.setInput() below will indirectly invoke _onEditorOpened. The other way
				// _onEditorOpened is invoked is through a jump-to-def. When a user does a jump-to-def, we need to
				// update the URL so that React/flux will fetch blob contents, etc. Doing this updates props.location,
				// and therefore this method to be invoked.
				//
				// We have the following cases to handle:
				// - initial page load: starts with _editorPropsChanged, we tell monaco where to move the cursor,
				//   (the second argument to this._editor.setInput below) and must not update the URL (doing so
				//   could make "official" a URL derived from an intermediate set of property values).
				// - jump-to-def: starts with _onEditorOpened, which calls router.push(url) and eventually invokes this
				//   method (at which point, we've already updated the URL and don't need to do so again)
				// - browser "back": starts with _editorPropsChanged (since props.location is updated), we simply
				//   tell monaco which uri to open and at what range. This is determined entirely by nextProps.location
				//   and we don't need _onEditorOpened to update the URL (the browser already did so).
				//
				// Jump-to-def starts with _onEditorOpened, and starting from that code path we ALWAYS update URL.
				// Therefore, whenever we invoke _onEditorOpened from _editorPropsChanged, we NEVER update URL.
				//
				// The reason that this._shortCircuitURLNavigationOnEditorOpened is an integer and not a boolean is
				// we might have more than one concurrent call to _editorPropsChanged. If we only stored a boolean,
				// there is a race condition where we treat one of the _onEditorOpened calls as being initiated by
				// a jump-to-def and change the URL as a result (this in causes a subtle bug where initial page loads
				// of links to blob lines redirect to the naked file URL instead of jumping to the correct line).
				// Note that to be pedantically correct, we should go even further to determine whether _onEditorOpened
				// calls originated from this code path or the jump-to-def path (we could use a stack of (uri, range)
				// tuples), but in practice, it's unlikely that any 2 of the above actions (initial page load,
				// jump-to-def, back button) would occur concurrently. So an integer is good enough.

				this._shortCircuitURLNavigationOnEditorOpened++; // when > 0, _onEditorOpened will not change the URL
				this._editor.setInput(uri, range).then(() => {
					// Always decrement this value after opening the editor.
					this._shortCircuitURLNavigationOnEditorOpened--;
					this._setEditorHighlightForLineSelection();
				});
			}
		}
	}

	_onKeyDownForFindInPage(e: KeyboardEvent & Event): void {
		// TODO(sqs): can make this unnecessary?
		const mac = navigator.userAgent.indexOf("Macintosh") >= 0;
		const ctrl = mac ? e.metaKey : e.ctrlKey;
		const FKey = 70;
		if (ctrl && (e.key === "KeyF" || e.keyCode === FKey)) {
			if (this._editor) {
				e.preventDefault();
				(document.getElementsByClassName("inputarea")[0] as any).focus();
				this._editor.trigger("keyboard", "actions.find", {});
			}
		}
	}

	_onLineSelected(): void {
		if (!this._editor || !this._editor.getSelection()) {
			return;
		}

		let selection = this._editor.getSelection();
		let startSelection;
		let startColumn;
		let endSelection;
		let endColumn;
		// Fix URL hash formatting to be in asc order.
		if (selection.selectionStartLineNumber < selection.positionLineNumber) {
			startSelection = selection.selectionStartLineNumber;
			startColumn = selection.selectionStartColumn;
			endSelection = selection.positionLineNumber;
			endColumn = selection.positionColumn;
		} else {
			startSelection = selection.positionLineNumber;
			startColumn = selection.positionColumn;
			endSelection = selection.selectionStartLineNumber;
			endColumn = selection.selectionStartColumn;
		}

		// Handle VSCode selection off by 1 case.
		if (startSelection !== endSelection) {
			if (selection.selectionStartColumn === 1 && selection.selectionStartLineNumber > selection.positionLineNumber) {
				endSelection = endSelection - 1;
			} else if (selection.endColumn === 1) {
				endSelection = endSelection - 1;
			}
		}

		let lineHash: string;
		if (startSelection !== endSelection) {
			lineHash = "#L" + startSelection + "-" + endSelection;
		} else {
			lineHash = "#L" + startSelection;
		}

		history.replaceState("", document.title, window.location.pathname + lineHash); // nice and clean
	}

	_onResize(e: Event): void {
		if (this._editor) {
			this._editor.layout();
		}
	}

	_setEditorHighlightForLineSelection(): void {
		if (this.props.startLine && this.props.startCol === null) {
			if (this._editor) {
				this._editor.setHighlightForLineSelection(this.props.startLine, this.props.endLine || this.props.startLine);
			}
		}
	}

	_onEditorOpened(e: IEditorOpenedEvent): void {
		if (this._shortCircuitURLNavigationOnEditorOpened > 0) {
			return;
		}

		let {repo, rev, path} = URIUtils.repoParams(e.model.uri);

		// If same repo, use the rev from the URL, so that we don't
		// change the address bar around a lot (bad UX).
		//
		// TODO(sqs): this will break true cross-rev same-repo jumps.
		if (repo === this.props.repo) {
			rev = this.props.rev;
		}

		let url = urlToBlob(repo, rev, path) + (this.props.location.search ? this.props.location.search : "");

		const sel = e.editor.getSelection();
		if (sel && (!sel.isEmpty() || sel.startLineNumber !== 1)) {
			let startCol: number | undefined = sel.startColumn;
			let endCol: number | undefined = sel.endColumn;
			if (e.model.getLineMinColumn(sel.startLineNumber) === startCol) {
				startCol = undefined;
			}
			if (e.model.getLineMaxColumn(sel.endLineNumber) === endCol) {
				endCol = undefined;
			}

			// HACK
			if (endCol <= 1 && !startCol) {
				endCol = undefined;
			}

			const r = RangeOrPosition.fromOneIndexed(sel.startLineNumber, startCol, sel.endLineNumber, endCol);
			url = `${url}#L${r.toString()}`;
		}

		this.context.router.push(url);
	}

	render(): JSX.Element | null {
		let title = trimRepo(this.props.repo);
		const pathParts = this.props.path ? this.props.path.split("/") : null;
		if (pathParts) {
			title = `${pathParts[pathParts.length - 1]} · ${title}`;
		}
		return (
			<FlexContainer direction="top_bottom" style={{flex:"auto", backgroundColor: colors.coolGray1()}}>
				<Helmet title={title} />
				<ChromeExtensionToast location={this.props.location}/>
				<OnboardingModals location={this.props.location}/>
				{this.props.location.query["tour"] && <TourOverlay location={this.props.location}/>}
				<BlobTitle
					repo={this.props.repo}
					path={this.props.path}
					repoObj={this.props.repoObj}
					rev={this.props.rev}
					commitID={this.props.commitID}
					routes={this.props.routes}
					routeParams={this.props.routeParams}
					isCloning={this.props.isCloning}
					toast={this.state.toast}
				/>
				<EditorComponent editorRef={this._setEditor} style={{ display: "flex", flex: "auto", width: "100%" }} />
			</FlexContainer>);
	}
}
