// tslint:disable: typedef ordered-imports

import {Location} from "history";
import * as React from "react";
import * as classNames from "classnames";

import {BlobLegacy} from "sourcegraph/blob/BlobLegacy";

import {BlobStore, keyForFile, keyForAnns} from "sourcegraph/blob/BlobStore";
import {BlobContentPlaceholder} from "sourcegraph/blob/BlobContentPlaceholder";
import {Container} from "sourcegraph/Container";
import {Store} from "sourcegraph/Store";
import {DefStore} from "sourcegraph/def/DefStore";
import {DefTooltip} from "sourcegraph/def/DefTooltip";
import {Link} from "react-router";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/blob/BlobBackend";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {lineFromByte} from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import * as styles from "sourcegraph/def/styles/Refs.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

const SNIPPET_REF_CONTEXT_LINES = 4; // Number of additional lines to show above/below a ref

interface Props {
	location: Location;
	repo?: string;
	rev?: string;
	commitID?: string;
	def?: string;
	defObj?: any;
	refs?: any;
	repoRefs: {
		Repo?: string;
		Files: any[];
	};
	refetch?: boolean;
	initNumSnippets?: number; // number of snippets to initially expand
	fileCollapseThreshold?: number; // number of files to show before "and X more..."-style paginator
	rangeLimit?: number;
	showRepoTitle?: boolean;
	refIndex?: number;
}

type State = any;

// The purpose of this file is to easily render one example during the onboarding flow that will allow the user to hover over examples
// in a sandboxed mode. This is similar to the regular RefsContainer except that css styles and eventlisteners / additional fetches for information
// are removed.
export class OnboardingExampleRefsContainer extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
	};

	rangesMemo: any;
	filesByName: any;
	ranges: any;
	anns: any;

	constructor(props: Props) {
		super(props);
		this.state = {
			shownFiles: new Set(),
			initExpanded: false, // Keep track of when we've auto-expanded snippets.
		};
		this.rangesMemo = {}; // optimization: cache the line range that should be displayed for each ref

		// optimization: these memos reduce the amount of component state which must be copied in reconcileState
		this.filesByName = {};
		this.ranges = {};
		this.anns = {};
	}

	shouldComponentUpdate(nextProps, nextState, nextContext) {
		if (super.shouldComponentUpdate(nextProps, nextState, nextContext)) {
			return true;
		}

		// Since the reference values of the memo'd state don't change even though the contents
		// may be updated (e.g. as a result of asynchronous fetches) reconcileState
		// must set a special flag if it resolves new data from a store which is kept in the memo.
		return Boolean(nextState.forceComponentUpdate);
	}

	stores(): Store<any>[] {
		return [DefStore, BlobStore];
	}

	reconcileState(state: State, props: Props): void {
		state.location = props.location || null;
		if (typeof state.showAllFiles === "undefined") {
			state.showAllFiles = false;
		}
		state.fileCollapseThreshold = props.fileCollapseThreshold || 3;

		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.showRepoTitle = props.showRepoTitle || false;

		state.refRepo = props.repoRefs.Repo || null;
		state.refRev = state.refRepo === state.repo ? state.rev : null;
		state.repoRefLocations = props.repoRefs || null;
		state.rangeLimit = props.rangeLimit || null;
		if (state.repoRefLocations) {
			state.fileLocations = state.repoRefLocations.Files;
		}

		state.refs = props.refs || DefStore.refs.get(state.repo, state.rev, state.def, state.refRepo, null);

		state.hoverInfos = DefStore.hoverInfos;
		state.hoverPos = DefStore.hoverPos;

		if (state.fileLocations && !state.initExpanded) {
			// Auto-expand N snippets by default.
			for (let i = 0; i < props.initNumSnippets; i++) {
				let loc = state.fileLocations[i];
				if (loc) {
					state.shownFiles.add(loc.Path);
				}
			}
			state.initExpanded = true;
		}

		state.forceComponentUpdate = false;
		if (state.refs && !state.refs.Error) {
			for (let ref of state.refs || []) {
				if (!ref) {
					continue;
				}
				let refRev = ref.CommitID;
				if (!this.filesByName[ref.File]) {
					let file = BlobStore.files[keyForFile(ref.Repo, refRev, ref.File)] || null;
					if (file) {
						// Pass through Error to this.filesByName (i.e., proceed even if file.Error is truthy).
						state.forceComponentUpdate = true;
						this.filesByName[ref.File] = file;
					}
				}

				if (this.filesByName[ref.File] && !this.filesByName[ref.File].Error) {
					this.ranges[ref.File] = this.ranges[ref.File] ? this.ranges[ref.File] : [];
					const rangeKey = `${ref.File}${ref.Start}`;
					if (!this.rangesMemo[rangeKey]) {
						// Response from LSP comes with StartLine, so we can directly use it if available.
						let startLineNum;
						let endLineNum;
						if (ref.StartLine) {
							startLineNum = ref.StartLine;
							endLineNum = ref.EndLine;
						} else {
							let contents = this.filesByName[ref.File].ContentsString;
							startLineNum = lineFromByte(contents, ref.Start);
							endLineNum = lineFromByte(contents, ref.End);
						}
						this.ranges[ref.File].push([
							Math.max(startLineNum - SNIPPET_REF_CONTEXT_LINES, 0),
							endLineNum + SNIPPET_REF_CONTEXT_LINES,
							startLineNum,
						]);
						this.rangesMemo[rangeKey] = true;
					}
				}
				if (!this.anns[ref.File]) {
					let anns = BlobStore.annotations[keyForAnns(ref.Repo, ref.CommitID, ref.File)] || null;
					if (anns) {
						// Pass through Error to this.anns (i.e., proceed even if anns.Error is truthy).
						state.forceComponentUpdate = true;
						this.anns[ref.File] = anns;
					}
				}
			}
		}
	}

	onStateTransition(prevState: State, nextState: State): void {
		const refPropsUpdated = prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def || prevState.refRepo !== nextState.refRepo;
		if (refPropsUpdated) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def, nextState.refRepo));
		}

		if (nextState.refs && nextState.refs.length > 0 && !nextState.refs.Error && (nextState.refs !== prevState.refs || nextState.shownFiles !== prevState.shownFiles)) {
			let firstRef = nextState.refs[0]; // hack: assuming that all refs given to a RefsContainer are from the same repo and rev, thus using the first ref to determine which files we want to show
			let repo = firstRef.Repo;
			let rev = firstRef.CommitID;
			for (let file of Array.from(nextState.shownFiles as Set<any>)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(repo, rev, file));
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(repo, rev, file));
			}
		}
	}

	_clickedFromRepo() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "FromRepoClicked", {repo: this.state.repo, def: this.state.def, ref_repo: this.state.refRepo});
	}

	render(): JSX.Element | null {
		return (
			<div className={classNames(base.pa4, colors.b__cool_pale_gray, styles.full_width_sm)}>
				<div className={styles.container}
					onMouseEnter={() => {
						if (!this.state.mouseover) {
							this.setState({mouseover: true, mouseout: false});
						}
					}}
					onMouseLeave={() => this.setState({mouseover: false, mouseout: true})}
					onMouseOut={() => Dispatcher.Stores.dispatch(new DefActions.Hovering(null))}>
					{/* mouseover state is for optimization which will only re-render the moused-over blob when a def is highlighted */}
					{/* this is important since there may be many ref containers on the page */}
					<div>
						<div>
							{this.state.fileLocations && this.state.fileLocations.map((loc, i) => {
								let err;
								let file = this.filesByName ? this.filesByName[loc.Path] : null;
								if (file && file.Error) {
									switch (file.Error.response.status) {
									case 413:
										err = "Sorry, this file is too large to display.";
										break;
									default:
										err = "File is not available.";
									}
								}
								if (this.state.refs && this.state.refs.Error) {
									err = `Error loading references for ${loc.Path}.`;
								}

								if (!file) {
									return <div key={i}><BlobContentPlaceholder key={i} numLines={SNIPPET_REF_CONTEXT_LINES * 2 + 1} /></div>;
								}

								let ranges = this.ranges[loc.Path];
								if (this.state.rangeLimit) {
									ranges = ranges.slice(0, this.state.rangeLimit);
									ranges.map((r) => [r[0], Math.min(r[0] + 10, r[1])]);
								}

								return (
									<div key={i} className={styles.single_ref_container}>
										<div className={styles.refs}>
											<BlobLegacy
												location={this.state.location}
												repo={this.state.refRepo}
												rev={this.state.refRev}
												commitID={this.state.commitID}
												path={loc.Path}
												contents={file.ContentsString}
												annotations={this.anns[loc.Path] || null}
												skipAnns={file.ContentsString && file.ContentsString.length >= 40 * 2500}
												activeDefRepo={this.state.repo}
												activeDef={this.state.def}
												lineNumbers={false}
												displayRanges={ranges || null}
												highlightedDef={null}
												highlightedDefObj={null}
												textSize="large" />
										</div>
										{this.state.refRepo && <div style={{textAlign: "center"}} className={classNames(colors.bg_light_blue, base.pv2, styles.f7, base.hidden_s)}>From <Link to={`${urlToBlob(this.state.refRepo, this.state.refRev, loc.Path)}${ranges ? `#L${ranges[0][2]}` : ""}`} onClick={this._clickedFromRepo.bind(this)}>{this.state.refRepo}</Link></div>}
									</div>
								);
							})}
						</div>
					</div>
					<DefTooltip
						currentRepo={this.state.repo}
						hoverPos={this.state.hoverPos}
						hoverInfos={this.state.hoverInfos} />
				</div>
			</div>
		);
	}
}
