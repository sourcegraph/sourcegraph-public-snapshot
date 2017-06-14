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

export interface DiffusionProps extends Props {
	repoURI: string;
	branch: string;
	rev: string;
}

export interface DifferentialProps extends Props {
	baseBranch: string;
	baseRepoURI: string;
	headBranch: string;
	headRepoURI: string;
}

interface State {
	resolvedRevs: { [key: string]: backend.ResolvedRevResp };
}

export abstract class PhabBlobAnnotator<P extends Props> extends React.Component<P, State> {
	revisionChecker: number;
	fileExtension: string;
	expandListenerAdded: boolean = false;

	constructor(props: Props) {
		super(props);
		this.state = {
			resolvedRevs: {},
		};
		this.fileExtension = utils.getPathExtension(props.path);
		this.callResolveRevs();
		// I noticed that on 1/5 of page loads, even though the annotations code was
		// successfully annotating elements, due to a timing thing the code elements on the page
		// were different than the ones the code had annotated (they'd been re-laid out or something).
		// As a quick fix, I re-call addAnnotations 2 seconds after page load.
		this.addAnnotations();
		this.addExpandListener();
		setTimeout(() => this.addAnnotations(), 2000);
		// for pages with large code diffs, sometimes it takes >5 seconds for the page to load
		setTimeout(() => this.addExpandListener(), 2000);
		setTimeout(() => this.addExpandListener(), 10000);
		// end of aforementioned hack
	}

	componentDidMount(): void {
		// Set a timer to re-check revision data every 5 seconds, for repos that haven't been
		// cloned and revs that haven't been sync'd to Sourcegraph.com.
		// Single-flighted requests / caching prevents spamming the API.
		this.revisionChecker = setInterval(() => this.callResolveRevs(), 3000);
	}

	private addExpandListener(): void {
		// javelin is hacked so that if a show-more link is clicked, this event is fired.
		const table = this.getTable();
		if (!table || this.expandListenerAdded) {
			return;
		}
		this.expandListenerAdded = true;
		table.addEventListener("expandClicked", () => {
			this.addAnnotations();
			setTimeout(() => this.addAnnotations(), 2000);
		});
	}

	componentWillUnmount(): void {
		if (this.revisionChecker) {
			clearInterval(this.revisionChecker);
		}
		const table = this.getTable();
		if (table) {
			table.removeEventListener("expandClicked");
		}
	}

	componentDidUpdate(): void {
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
		}).catch(() => {
			// no-op. we only want to print errors once, they are printed during promise creations
		});
	}

	applyAnnotationsIfResolvedRev(uri: string, isBase: boolean, rev?: string): void {
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
		if (resolvedRev && resolvedRev.commitID) {
			const cells = this.getCodeCells(isBase);
			addAnnotations(this.props.path, { repoURI: uri, rev: resolvedRev.commitID, isDelta: true, isBase: isBase }, cells);
		}
	}

	getTable(): HTMLTableElement | null {
		return this.props.blobElement.querySelector("table");
	}

	protected getFileOpenCallback = (): void => {
		eventLogger.logOpenFile(this.getEventLoggerProps());
	}

	abstract addAnnotations(): void;
	abstract getCodeCells(isBase: boolean): CodeCell[];
	abstract callResolveRevs(): void;
	abstract getEventLoggerProps(): any;
}

const iconStyle = { marginTop: "-1px", paddingRight: "4px", fontSize: "18px", height: ".8em", width: ".8em" };
export function SourcegraphButton(blobUrl: string, classNames: string, eventHandler: () => void): JSX.Element {
	// TODO(john): consolidate w/ other blob annotators (and bring back auth-required grayscale)
	return (
		<a title="View in Sourcegraph" className={classNames} href={blobUrl} onClick={() => eventHandler()}><SourcegraphIcon style={iconStyle} />
			<span className="sg-clickable">Sourcegraph</span>
		</a>
	);
}
