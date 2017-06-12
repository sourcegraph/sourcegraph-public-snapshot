import * as backend from "../backend";
import * as utils from "../utils";
import { sourcegraphUrl } from "../utils/context";
import * as phabricator from "../utils/phabricator";
import { CodeCell } from "../utils/types";
import { DiffusionProps, PhabBlobAnnotator, SourcegraphButton } from "./PhabBlobAnnotator";

export class PhabDiffusionBlobAnnotator extends PhabBlobAnnotator<DiffusionProps> {
	constructor(props: DiffusionProps) {
		super(props);
	}

	addAnnotations(): void {
		this.applyAnnotationsIfResolvedRev(this.props.repoURI, false, this.props.rev);
	}

	getEventLoggerProps(): any {
		return {
			repo: this.props.repoURI,
			path: this.props.path,
			language: this.fileExtension,
		};
	}

	callResolveRevs(): void {
		this.resolveRevs(this.props.repoURI, this.props.rev);
	}

	getCodeCells(isBase: boolean): CodeCell[] {
		const table = this.getTable();
		if (!table) {
			return [];
		}
		return phabricator.getCodeCellsForAnnotation(table);
	}

	render(): JSX.Element | null {
		const DIFFUSION_CLASSES = "button grey has-icon msl phui-header-action-link";
		if (!this.state.resolvedRevs[backend.cacheKey(this.props.repoURI, this.props.rev)]) {
			return null;
		}
		return SourcegraphButton(
			utils.getSourcegraphBlobUrl(sourcegraphUrl, this.props.repoURI, this.props.path, this.props.rev),
			this.props.repoURI,
			DIFFUSION_CLASSES,
			this.getFileOpenCallback,
		);
	}
}
