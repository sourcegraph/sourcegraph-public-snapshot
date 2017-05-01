import * as React from "react";
import * as backend from "../backend";
import * as utils from "../utils";
import { addAnnotations } from "../utils/annotations";
import { eventLogger } from "../utils/context";
import { CodeCell } from "../utils/types";
import { SourcegraphIcon } from "./Icons";
import * as _ from "lodash";

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

	fileExtension: string;

	// revisionChecker is a timer used to sync the current repo/revision to this component
	revisionChecker: NodeJS.Timer;

	scrollCallback?: () => void;

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
		this.revisionChecker = setInterval(() => this.resolveRevs(this.props.repo, this.props.rev), 5000);

		// scrollCallback
		this.scrollCallback = _.debounce(() => this.addAnnotations(), 500, { leading: true });
		document.addEventListener("scroll", this.scrollCallback);
	}

	componentWillUnmount(): void {
		if (this.scrollCallback) {
			document.removeEventListener("scroll", this.scrollCallback);
			this.scrollCallback = undefined;
		}
		if (this.revisionChecker) {
			clearInterval(this.revisionChecker);
		}
	}

	componentDidUpdate(prevProps: Props, prevState: State): void {
		// if the state changes, it means we resolved a rev. the loading button will change, and we need to add annotations.
		this.addAnnotations();
	}

	// resolveRevs resolves the specified revision of the repository and then sets both on the component state.
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

	// addAnnotationsIfResolvedRev adds annotations to the DOM if the revision has been properly resolved.
	// It is idempotent, so it can be called multiple times, and for Bitbucket Server, it should be called
	// multiple times as the DOM changes as the user scrolls.
	addAnnotationsIfResolvedRev(uri: string, isBase: boolean, rev?: string): void {
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

	addAnnotations = (): void => {
		this.addAnnotationsIfResolvedRev(this.props.repo, false, this.props.rev);
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
		return null;
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
