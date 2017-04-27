import * as React from "react";
import * as backend from "../backend";
import * as utils from "../utils";
import { addAnnotations } from "../utils/annotations";
import { eventLogger } from "../utils/context";
import { CodeCell } from "../utils/types";
import { SourcegraphIcon } from "./Icons";

interface Props {
	blobElement: HTMLElement;
	path: string;
}

export interface BitbucketBrowseProps extends Props {
	projectCode: string;
	repo: string;
	rev: string;
}

interface State {
	resolvedRevs: { [key: string]: backend.ResolvedRevResp };
}

export class BitbucketBlobAnnotator extends React.Component<BitbucketBrowseProps, State> {
	revisionChecker: NodeJS.Timer;
	fileExtension: string;
	expandListenerAdded: boolean = false;
	scrollTimer: NodeJS.Timer | null;

	constructor(props: BitbucketBrowseProps) {
		super(props);
		this.state = {
			resolvedRevs: {},
		};
		this.fileExtension = utils.getPathExtension(props.path);
		this.resolveRevs(this.props.repo, this.props.rev);

		// I noticed that on 1/5 of page loads, even though the annotations code was
		// successfully annotating elements, due to a timing thing the code elements on the page
		// were different than the ones the code had annotated (they'd been re-laid out or something).
		// As a quick fix, I re-call addAnnotations 2 seconds after page load.
		this.addAnnotations();
		setTimeout(() => {
			this.addAnnotations();
		}, 2000);
	}

	componentDidMount(): void {
		// Set a timer to re-check revision data every 5 seconds, for repos that haven't been
		// cloned and revs that haven't been sync'd to Sourcegraph.com.
		// Single-flighted requests / caching prevents spamming the API.
		this.revisionChecker = setInterval(() => this.resolveRevs(this.props.repo, this.props.rev), 3000);
		document.addEventListener("scroll", this.scrollCallback);
	}

	scrollCallback = (): void => {
		if (this.scrollTimer) {
			clearTimeout(this.scrollTimer);
		}
		this.scrollTimer = setTimeout(this.addAnnotations, 500);
	}

	componentWillUnmount(): void {
		if (this.revisionChecker) {
			clearInterval(this.revisionChecker);
		}
		if (this.scrollTimer) {
			clearInterval(this.scrollTimer);
		}
	}

	componentDidUpdate(prevProps: Props, prevState: State): void {
		// if the state changes, it means we resolved a rev. the loading button will change, and we need to add annotations.
		this.addAnnotations();
	}

	resolveRevs(repo: string, rev: string): void {
		const key = backend.cacheKey(repo, rev);
		if (this.state.resolvedRevs[key] && this.state.resolvedRevs[key].notFound) {
			// User doesn't have permission to view repo; no need to fetch.
			return;
		}
		if (this.state.resolvedRevs[key] && this.state.resolvedRevs[key].commitID) {
			return; // nothing to do, because repo has already been resolved.
		}
		backend.resolveRev(repo, rev).then((resp) => {
			let repoStat;
			if (rev) {
				// Empty rev is checked to determine if the user has access to the repo.
				// Non-empty is checked to determine if Sourcegraph.com is sync'd.
				repoStat = { [repo]: resp };
			}
			this.setState({ resolvedRevs: Object.assign({}, this.state.resolvedRevs, { [key]: resp }, repoStat) });
		});
	}

	applyAnnotationsIfResolvedRev(uri: string, isBase: boolean, rev?: string): void {
		if (!utils.supportedExtensions.has(this.fileExtension)) {
			return; // Don't annotate unsupported languages
		}
		// this is outside of the resolveRev area, becuase this is asking if the view changed
		// and is less concerned with if we happened to have annotations. I figure it's safer
		// to put this code outside of that if, to avoid the 1 second poller overwhelming the page
		const table = this.getTable();
		if (!table) {
			return;
		}
		// switching file views blows away the table, and on differential views we take advantage of this by noticing the dropped class
		table.classList.add("sg-table-annotated");
		const resolvedRev = this.state.resolvedRevs[backend.cacheKey(uri, rev)];
		const ext = utils.getPathExtension(this.props.path);
		const spacesToTab = Boolean(ext) && ext === "go" ? 4 : 0;
		if (resolvedRev && resolvedRev.commitID) {
			const cells = this.getCodeCells(isBase);
			addAnnotations(this.props.path, { repoURI: uri, rev: resolvedRev.commitID, isDelta: false, isBase: isBase }, table, this.getEventLoggerProps(), cells, spacesToTab);
		}
	}

	getTable(): HTMLTableElement | null {
		return this.props.blobElement.querySelector(".CodeMirror-lines") as HTMLTableElement | null;
	}

	protected getFileOpenCallback = (): void => {
		eventLogger.logOpenFile(this.getEventLoggerProps());
	}

	addAnnotations = (): void => {
		this.applyAnnotationsIfResolvedRev(this.props.repo, false, this.props.rev);
	}

	getEventLoggerProps(): Object {
		return {
			repo: this.props.repo,
			path: this.props.path,
			language: this.fileExtension,
		};
	}

	getCodeCells(isBase: boolean): CodeCell[] {
		const table = this.getTable();
		if (!table) {
			return [];
		}
		return getCodeCellsForAnnotation(table);
	}

	render(): JSX.Element | null {
		if (!this.state.resolvedRevs[backend.cacheKey(this.props.repo, this.props.rev)]) {
			return null;
		}
		return SourcegraphButton(
			utils.supportedExtensions.has(this.fileExtension),
			this.state.resolvedRevs[this.props.repo].cloneInProgress as boolean,
			"http://node.aws.sgdev.org:30000/github.com/gorilla/mux/-/blob/mux.go",
			this.props.repo,
			utils.upcomingExtensions.has(this.fileExtension),
			"",
			this.getFileOpenCallback,
		);
	}
}

/**
 * getCodeCellsForAnnotation code cells which should be annotated
 */
export function getCodeCellsForAnnotation(table: HTMLTableElement): CodeCell[] {
	const code = table.getElementsByClassName("CodeMirror-code")[0];
	const cells: CodeCell[] = [];
	// tslint:disable-next-line:prefer-for-of
	let count = 1;

	const children = Array.from(code.children);
	if (children && children.length > 0 && children[0].getElementsByClassName("line-number")[0]) {
		count = parseInt(children[0].getElementsByClassName("line-number")[0].innerText, 10);
	}

	for (const row of children) {
		const element = row.getElementsByClassName("CodeMirror-line")[0];
		cells.push({
			cell: element as HTMLElement,
			line: count,
			isAddition: false,
			isDeletion: false,
		});
		count++
	}
	return cells;
}

const iconStyle = { marginTop: "-1px", paddingRight: "4px", fontSize: "18px", height: ".8em", width: ".8em" };
const disabledStyle = { cursor: "default", opacity: 0.6 };
export function SourcegraphButton(isFileSupported: boolean, isLoading: boolean, blobUrl: string, repoName: string, comingSoon: boolean, classNames: string, eventHandler: () => void): JSX.Element {
	if (!isFileSupported) {
		const tooltipLabel = !comingSoon ? "File not supported" : "Language support coming soon!";
		return (<a className={classNames} title={tooltipLabel} style={disabledStyle}>
			<SourcegraphIcon style={iconStyle} />
			Sourcegraph
			</a>);

	} else if (isLoading) {
		return (<a className={classNames} title={`Sourcegraph is analyzing ${repoName}`} style={disabledStyle}>
			<SourcegraphIcon style={iconStyle} />
			Loading...
			</a>);
	} else {
		return (
			<a title="View in Sourcegraph" className={classNames} href={blobUrl} onClick={() => eventHandler()}><SourcegraphIcon style={iconStyle} />
				<span className="sg-clickable">Sourcegraph</span>
			</a>
		);
	}
}
