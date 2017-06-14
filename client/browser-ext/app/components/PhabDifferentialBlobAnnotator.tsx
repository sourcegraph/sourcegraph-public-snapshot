import * as backend from "../backend";
import * as utils from "../utils";
import { sourcegraphUrl } from "../utils/context";
import * as phabricator from "../utils/phabricator";
import { CodeCell } from "../utils/types";
import { DifferentialProps, PhabBlobAnnotator, SourcegraphButton } from "./PhabBlobAnnotator";

export class PhabDifferentialBlobAnnotator extends PhabBlobAnnotator<DifferentialProps> {
	private viewChangedChecker: number;

	constructor(props: DifferentialProps) {
		super(props);
	}
	/**
	 * Phabricator rips out the table element when you change the view, but leaves the containing blob.
	 * We need to check the table at regular intervals to try and detect this.
	 * This is separate from context expansion, which is detected by this.addExpandListener().
	 */
	viewChanged(): boolean {
		const table = this.getTable();
		if (!table) {
			// this means the table hasn't loaded, not worth triggering an annotation
			return false;
		}
		return !table.classList.contains("sg-table-annotated");
	}

	isSplitView(): boolean {
		const table = this.getTable();
		if (!table) {
			return false; // false by default because this page loads with unified default
		}
		return table.classList.contains("diff-2up");
	}

	triggerAddAnnotationsIfViewChanged(): void {
		if (this.viewChanged()) {
			this.addAnnotations();
		}
	}

	componentDidMount(): void {
		super.componentDidMount();
		this.viewChangedChecker = setInterval(() => this.triggerAddAnnotationsIfViewChanged(), 3000);
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();
		if (this.viewChangedChecker) {
			clearInterval(this.viewChangedChecker);
		}
	}

	addAnnotations(): void {
		this.applyAnnotationsIfResolvedRev(this.props.headRepoURI, false, this.props.headBranch);
		this.applyAnnotationsIfResolvedRev(this.props.baseRepoURI, true, this.props.baseBranch);
	}

	getEventLoggerProps(): any {
		return {
			repo: this.props.headRepoURI,
			path: this.props.path,
			language: this.fileExtension,
			isPullRequest: true,
			isCommit: false,
		};
	}

	callResolveRevs(): void {
		this.resolveRevs(this.props.headRepoURI, this.props.headBranch);
		this.resolveRevs(this.props.baseRepoURI, this.props.baseBranch);
	}

	getCodeCells(isBase: boolean): CodeCell[] {
		const table = this.getTable();
		if (!table) {
			return [];
		}
		return phabricator.getCodeCellsForDifferentialAnnotations(table, this.isSplitView(), isBase);
	}

	render(): JSX.Element | null {
		const DIFFERENTIAL_CLASSES = "button grey has-icon msl";
		if (!this.state.resolvedRevs[backend.cacheKey(this.props.baseRepoURI, this.props.baseBranch)]) {
			return null;
		}
		return SourcegraphButton(
			utils.getSourcegraphBlobUrl(sourcegraphUrl, this.props.headRepoURI, this.props.path, this.props.headBranch),
			DIFFERENTIAL_CLASSES,
			this.getFileOpenCallback,
		);
	}
}
