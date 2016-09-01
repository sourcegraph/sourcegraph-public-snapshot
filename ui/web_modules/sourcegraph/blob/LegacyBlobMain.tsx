// tslint:disable: typedef ordered-imports

import {Location} from "history";
import * as React from "react";
import Helmet from "react-helmet";

import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Store} from "sourcegraph/Store";
import {BlobLegacy} from "sourcegraph/blob/BlobLegacy";
import {BlobContentPlaceholder} from "sourcegraph/blob/BlobContentPlaceholder";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobToolbar} from "sourcegraph/blob/BlobToolbar";
import {FileMargin} from "sourcegraph/blob/FileMargin";
import {DefTooltip} from "sourcegraph/def/DefTooltip";
import {DefStore} from "sourcegraph/def/DefStore";
import "sourcegraph/blob/BlobBackend";
import "sourcegraph/def/DefBackend";
import "sourcegraph/build/BuildBackend";
import * as Style from "sourcegraph/blob/styles/Blob.css";
import {lineCol, lineRange, parseLineRange} from "sourcegraph/blob/lineCol";
import {urlTo} from "sourcegraph/util/urlTo";
import {makeRepoRev, trimRepo} from "sourcegraph/repo/index";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";
import {Header} from "sourcegraph/components/Header";
import {createLineFromByteFunc} from "sourcegraph/blob/lineFromByte";
import {defTitle, defTitleOK} from "sourcegraph/def/Formatter";

interface Props {
	repo: string;
	rev?: string;
	commitID?: string;
	path?: string;
	blob?: any;
	anns?: any;
	def?: any;
	skipAnns?: boolean;
	startLine?: number;
	startCol?: number;
	startByte?: number;
	endLine?: number;
	endCol?: number;
	endByte?: number;
	location: Location;
	children?: React.ReactNode;
}

type State = any;

export class LegacyBlobMain extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	_dispatcherToken: string;

	constructor(props: Props) {
		super(props);
		this.state = {
			selectionStartLine: null,
		};

	}

	componentDidMount(): void {
		super.componentDidMount();
		this._dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();
		Dispatcher.Stores.unregister(this._dispatcherToken);
	}

	reconcileState(state: State, props: Props): void {
		state.repo = props.repo;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.path = props.path || null;
		state.blob = props.blob || null;
		state.anns = props.anns || null;
		state.skipAnns = props.skipAnns || false;
		state.startLine = props.startLine || null;
		state.startCol = props.startCol || null;
		state.startByte = props.startByte || null;
		state.endLine = props.endLine || null;
		state.endCol = props.endCol || null;
		state.endByte = props.endByte || null;
		state.def = props.def || null;
		state.defObj = state.def && state.commitID ? DefStore.defs.get(state.repo, state.commitID, state.def) : null;
		state.children = props.children || null;

		state.hoverInfos = DefStore.hoverInfos;
		state.hoverPos = DefStore.hoverPos;
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState.blob !== nextState.blob) {
			nextState.lineFromByte = nextState.blob && typeof nextState.blob.ContentsString !== "undefined" ? createLineFromByteFunc(nextState.blob.ContentsString) : null;
		}
	}

	stores(): Store<any>[] {
		return [DefStore];
	}

	__onDispatch(action) {
		if (action instanceof BlobActions.SelectLine) {
			this._navigate(action.repo, action.rev, action.path, action.line ? `L${action.line}` : null);
		} else if (action instanceof BlobActions.SelectLineRange) {
			let pos = (this.props.location as any).hash ? parseLineRange((this.props.location as any).hash.replace(/^#L/, "")) : null;
			const startLine = Math.min(pos ? pos.startLine : action.line, action.line);
			const endLine = Math.max(pos ? (pos.endLine || pos.startLine) : action.line, action.line);
			this._navigate(action.repo, action.rev, action.path, startLine && endLine ? `L${lineRange(startLine, endLine)}` : null);
		} else if (action instanceof BlobActions.SelectCharRange) {
			let hash = action.startLine ? `L${lineRange(lineCol(action.startLine, action.startCol), action.endLine && lineCol(action.endLine, action.endCol))}` : null;
			this._navigate(action.repo, action.rev, action.path, hash);
		}
	}

	_navigate(repo, rev, path, hash) {
		let url = urlTo("blob", {splat: [makeRepoRev(repo, rev), path]} as any);

		// Replace the URL if we're just changing the hash. If we're changing
		// more (e.g., from a def URL to a blob URL), then push.
		const replace = (this.props.location as any).pathname === url;
		if (hash) {
			url = `${url}#${hash}`;
		}
		if (replace) {
			(this.context as any).router.replace(url);
		} else {
			(this.context as any).router.push(url);
		}
	}

	render(): JSX.Element | null {
		if (this.state.blob && this.state.blob.Error) {
			let msg;
			switch (this.state.blob.Error.response.status) {
			case 413:
				msg = "Sorry, this file is too large to display.";
				break;
			default:
				msg = "File is not available.";
			}
			return (
				<Header
					title={`${httpStatusCode(this.state.blob.Error)}`}
					subtitle={msg} />
			);
		}

		// NOTE: Title should be kept in sync with app/internal/ui in Go.
		let title = trimRepo(this.state.repo);
		const pathParts = this.state.path ? this.state.path.split("/") : null;
		if (pathParts) {
			title = `${pathParts[pathParts.length - 1]} · ${title}`;
		}
		if (this.state.defObj && !this.state.defObj.Error && defTitleOK(this.state.defObj)) {
			title = `${defTitle(this.state.defObj)} · ${title}`;
		}
		return (
			<div className={Style.container}>
				{title && <Helmet title={title} />}
				<div className={Style.spacer} />
				<div className={Style.blobAndToolbar}>
					<BlobToolbar
						repo={this.state.repo}
						rev={this.state.rev}
						commitID={this.state.commitID}
						path={this.state.path} />
					{(!this.state.blob || (this.state.blob && !this.state.blob.Error && !this.state.skipAnns && !this.state.anns)) && <BlobContentPlaceholder />}
					{this.state.blob && !this.state.blob.Error && typeof this.state.blob.ContentsString !== "undefined" && (this.state.skipAnns || (this.state.anns && !this.state.anns.Error)) &&
					<BlobLegacy
						startlineCallback = {node => this.setState({selectionStartLine: node})}
						location={this.props.location}
						repo={this.state.repo}
						rev={this.state.rev}
						commitID={this.state.commitID}
						path={this.state.path}
						contents={this.state.blob.ContentsString}
						annotations={this.state.anns}
						skipAnns={this.state.skipAnns}
						lineNumbers={true}
						highlightSelectedLines={true}
						highlightedDef={null}
						highlightedDefObj={null}
						activeDef={this.state.def}
						startLine={this.state.startLine}
						startCol={this.state.startCol}
						startByte={this.state.startByte}
						endLine={this.state.endLine}
						endCol={this.state.endCol}
						endByte={this.state.endByte}
						scrollToStartLine={true}
						dispatchSelections={true} />}
					<DefTooltip
						currentRepo={this.state.repo}
						hoverPos={this.state.hoverPos}
						hoverInfos={this.state.hoverInfos} />
				</div>
				<FileMargin
					className={Style.margin}
					style={(!this.state.blob || !this.state.anns) ? {visibility: "hidden"} : {}}
					lineFromByte={this.state.lineFromByte}
					selectionStartLine={this.state.selectionStartLine ? this.state.selectionStartLine : null}
					startByte={this.state.startByte}>
					{this.state.children}
				</FileMargin>
			</div>
		);
	}
}
