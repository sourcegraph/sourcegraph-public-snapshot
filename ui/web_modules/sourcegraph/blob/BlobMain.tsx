// tslint:disable: typedef ordered-imports

import {Location} from "history";
import * as React from "react";
import Helmet from "react-helmet";

import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Store} from "sourcegraph/Store";
import {Blob} from "sourcegraph/blob/Blob";
import * as DefActions from  "sourcegraph/def/DefActions";
import {DefStore} from "sourcegraph/def/DefStore";
import "sourcegraph/blob/BlobBackend";
import "sourcegraph/def/DefBackend";
import "sourcegraph/build/BuildBackend";
import * as Style from "sourcegraph/blob/styles/Blob.css";
import {urlTo} from "sourcegraph/util/urlTo";
import {makeRepoRev, trimRepo} from "sourcegraph/repo";
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

export class BlobMain extends Container<Props, State> {
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
		document.body.style.overflowY = "hidden";
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();
		Dispatcher.Stores.unregister(this._dispatcherToken);
		document.body.style.overflowY = "auto";
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
		state.rev = props.rev || "master";

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
		if (action instanceof DefActions.JumpDefFetched) {
			if (action.def.Error) {
				console.log("Go-to-definition failed:", action.def.Error); // tslint:disable-line
			} else {
				const rev = this.props.rev ? action.commitID : "";
				const url = urlToDef(action.def, rev);
				(this.context as any).router.push(url);
			}
		}
	}

	_navigate(repo: string, rev: string, path: string, hash: string): void {
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
			<div className={Style.container_monaco}>
				<Helmet title={title} />
				{this.state.blob && typeof this.state.blob.ContentsString === "string" && <Blob
					repo={this.state.repo}
					rev={this.state.rev}
					path={this.state.path}
					contents={this.state.blob.ContentsString}
					startLine={this.state.startLine}
					endLine={this.state.endLine}
					startByte={this.state.startByte} />}
			</div>
		);
	}
}
