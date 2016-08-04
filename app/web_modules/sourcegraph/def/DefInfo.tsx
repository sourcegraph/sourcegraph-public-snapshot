// tslint:disable

import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";
import "sourcegraph/blob/BlobBackend";
import "whatwg-fetch";

import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import base from "sourcegraph/components/styles/_base.css";
import typography from "sourcegraph/components/styles/_typography.css";

import annotationsByLine from "sourcegraph/blob/annotationsByLine";
import BlobLine from "sourcegraph/blob/BlobLine";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore, {keyForFile, keyForAnns} from "sourcegraph/blob/BlobStore";
import blobStyles from "sourcegraph/blob/styles/Blob.css";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import DefStore from "sourcegraph/def/DefStore";
import RepoRefsContainer from "sourcegraph/def/RepoRefsContainer";
import ExamplesContainer from "sourcegraph/def/ExamplesContainer";
import * as DefActions from "sourcegraph/def/DefActions";
import fileLines from "sourcegraph/util/fileLines";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {urlToDef} from "sourcegraph/def/routes";
import {qualifiedNameAndType, defTitle, defTitleOK} from "sourcegraph/def/Formatter";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {trimRepo} from "sourcegraph/repo/index";
import {urlToRepo} from "sourcegraph/repo/routes";
import {EmptyNodeIllo} from "sourcegraph/components/symbols/index";
import {Header, Heading, FlexContainer, GitHubAuthButton, Loader} from "sourcegraph/components/index";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

// Number of characters of the Docstring to show before showing the "collapse" options.
const DESCRIPTION_CHAR_CUTOFF = 500;
//
class DefInfo extends Container<any, any> {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
	};

	static propTypes = {
		repo: React.PropTypes.string,
		repoObj: React.PropTypes.object,
		def: React.PropTypes.string.isRequired,
		commitID: React.PropTypes.string,
		rev: React.PropTypes.string,
		defObj: React.PropTypes.object,
	};

	constructor(props) {
		super(props);
		this.state = {
			defDescrHidden: null,
		};
		this.splitHTMLDescr = this.splitHTMLDescr.bind(this);
		this.splitPlainDescr = this.splitPlainDescr.bind(this);
		this._onViewMore = this._onViewMore.bind(this);
		this._onViewLess = this._onViewLess.bind(this);
	}

	stores() {
		return [DefStore, BlobStore];
	}

	componentDidMount() {
		if (super.componentDidMount) super.componentDidMount();
		// Fix a bug where navigating from a blob page here does not cause the
		// browser to scroll to the top of this page.
		if (typeof window !== "undefined") window.scrollTo(0, 0);
	}

	shouldHideDescr(defObj, cutOff: number) {
		if (defObj.DocHTML) {
			let parser = new DOMParser();
			let doc = parser.parseFromString(defObj.DocHTML.__html, "text/html");
			return doc.documentElement.textContent && doc.documentElement.textContent.length >= cutOff;
		} else if (defObj.Docs) {
			return defObj.Docs[0].Data.length >= cutOff;
		}
		return false;
	}

	splitHTMLDescr(html, cutOff) {
		let parser = new DOMParser();
		let doc = parser.parseFromString(html, "text/html");

		// lp recreates the HTML tree by doing a DFS traversal, keeping
		// track of the consumed characters along the way
		// lp breaks early if the # of consumed characters exceeds our cutoff
		function lp(node, oldLength) {
			let childrenCopy: any[] = [];
			while (node.firstChild) {
				let clone = node.firstChild.cloneNode(true);
				childrenCopy.push(clone);
				node.removeChild(node.firstChild);
			}

			let newLength = node.textContent.length + oldLength;
			if (newLength >= cutOff) {
				node.textContent = `${node.textContent.slice(0, cutOff - node.textContent.length - 1)}...`;
				return [node, cutOff];
			}

			let latestLength = newLength;
			for (let child of childrenCopy) {
				let [newChild, newCount] = lp(child, latestLength);
				node.appendChild(newChild);
				latestLength = newCount;
				if (latestLength >= cutOff) {
					break;
				}
			}
			return [node, latestLength];
		}
		let newDoc = lp(doc.documentElement, 0)[0];
		return newDoc.getElementsByTagName("body")[0].innerHTML;
	}

	splitPlainDescr(txt, cutOff) {
		return txt.slice(0, Math.min(txt.length, cutOff));
	}

	_onViewMore() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedViewMoreDescription", {repo: this.state.repo, def: this.state.def, num_examples: this.state.examples["RepoRefs"].length});
		this.setState({defDescrHidden: false});
	}

	_onViewLess() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedViewLessDescription", {repo: this.state.repo, def: this.state.def});
		this.setState({defDescrHidden: true});
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.defCommitID = props.defObj ? props.defObj.CommitID : null;
		state.authors = state.defObj ? DefStore.authors.get(state.repo, state.defObj.CommitID, state.def) : null;
		state.refLocations = state.def && state.defObj ? DefStore.getRefLocations({
			repo: state.repo, commitID: state.defCommitID, def: state.def, repos: [],
		}) : null;
		state.examples = state.def && state.defObj ? DefStore.getExamples({
			repo: state.repo, commitID: state.defCommitID, def: state.def,
		}) : null;

		if (state.defObj && state.defDescrHidden === null) {
			state.defDescrHidden = this.shouldHideDescr(state.defObj, DESCRIPTION_CHAR_CUTOFF);
		}
	}

	onStateTransition(prevState, nextState) {
		if (prevState.defCommitID !== nextState.defCommitID && nextState.defCommitID) {
			Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.defCommitID, nextState.def));
		}
		if (nextState.currPage !== prevState.currPage || nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def || nextState.defObj !== prevState.defObj) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
				repo: nextState.repo, commitID: nextState.defCommitID, def: nextState.def, repos: nextState.defRepos, page: nextState.currPage,
			}));
		}
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def || nextState.defObj !== prevState.defObj) {
			Dispatcher.Backends.dispatch(new DefActions.WantExamples({
				repo: nextState.repo, commitID: nextState.defCommitID, def: nextState.def,
			}));
		}
		if (nextState.defObj !== null && nextState.defObj !== prevState.defObj) {
			Dispatcher.Backends.dispatch(new BlobActions.WantFile(
				nextState.defObj.Repo,
				nextState.defObj.CommitID,
				nextState.defObj.File,
			));
			Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(
				nextState.defObj.Repo,
				nextState.defObj.CommitID,
				nextState.defObj.File,
			));
		}
	}

	_getDefLine(defObj) {
		if (!defObj) {
			return null;
		}

		let file = BlobStore.files[keyForFile(defObj.Repo, defObj.CommitID, defObj.File)] || null;
		let anns = BlobStore.annotations[keyForAnns(defObj.Repo, defObj.CommitID, defObj.File)] || null;
		if (!file || !anns) {
			return null;
		}

		let lines = fileLines(file.ContentsString);
		let startLine = lineFromByte(lines, defObj.DefStart) - 1;
		let contents = lines[startLine];
		let lineAnns = annotationsByLine(anns.LineStartBytes, anns.Annotations, lines)[startLine];

		if (defObj.Kind !== "func") {
			return null;
		}

		// try to remove dangling open curly brace at the end of the line,
		// and detect if this single-line rendering method would fail
		const whiteSpace = /\s/;
		for (let i = contents.length - 1; i >= 0; i--) {
			if (!whiteSpace.test(contents[i])) {
				if (contents[i] !== "{") { i++; }
				let trimmedLine = contents.substr(0, i).trim();
				let typeParts = defObj.FmtStrings.Type.ScopeQualified.split(/[\s\.\/]/);
				let lastTypePart = typeParts.pop();
				if (!trimmedLine.endsWith(lastTypePart)) {
					// if the trimmed line doesn't end with the last token of the type,
					// we're dealing with a tricky function header -> abort
					return null;
				}
				contents = `${contents.substr(0, i)} ${contents.substr(i + 1)}`;
				break;
			}
		}

		return {
			contents: contents,
			anns: lineAnns,
			startByte: anns.LineStartBytes[startLine],
		};
	}

	_viewDefinitionClicked() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedViewDefinition", {repo: this.state.repo, def: this.state.def, num_examples: this.state.examples["RepoRefs"].length});
	}

	_viewRepoClicked() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedRepoAboveExamples", {repo: this.state.repo, def: this.state.def, num_examples: this.state.examples["RepoRefs"].length});
	}

	render(): JSX.Element | null {
		const {defObj, defDescrHidden, refLocations, examples, repo, rev, defCommitID, def} = this.state;
		let defBlobUrl = defObj ? urlToDef(defObj, rev) : "";

		if (refLocations && refLocations.Error) {
			return (
				<Header
					title={`${httpStatusCode(refLocations.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		let defLine = this._getDefLine(defObj);

		return (
			<FlexContainer styleName="bg_cool_pale_gray_2 flex_grow">
				<div styleName="container_fixed" className={base.mv3}>
					{/* NOTE: This should (roughly) be kept in sync with page titles in app/internal/ui. */}
					<Helmet title={defTitleOK(defObj) ? `${defTitle(defObj)} · ${trimRepo(repo)}` : trimRepo(repo)} />
					{defObj &&
						<div className={`${base.mv4} ${base.ph4}`}>
							<Heading level="5" styleName="break_word" className={base.mv2}>
								<table>
									<tbody>
										{defLine &&
											<BlobLine
												clickEventLabel="DefTitleTokenClicked"
												ref={null}
												repo={repo}
												rev={rev}
												commitID={defCommitID}
												path={defObj.Path}
												lineNumber={null}
												showLineNumber={false}
												startByte={defLine.startByte}
												contents={defLine.contents}
												textSize="deftitle"
												lineContentClassName={blobStyles.defTitleLineContent}
												annotations={defLine.anns}
												selected={null}
												highlightedDef={null}
												highlightedDefObj={null}
												activeDef={null}
												activeDefRepo={defObj.Repo} />
										}
										{!defLine &&
											<tr className={`${blobStyles.line} ${blobStyles.deftitle}`}>
												<td className={`code ${blobStyles.defTitleLineContent}`}>
													{qualifiedNameAndType(defObj, {unqualifiedNameClass: styles.def})}
												</td>
											</tr>
										}
									</tbody>
								</table>
							</Heading>

							{/* TODO DocHTML will not be set if the this def was loaded via the
								serveDefs endpoint instead of the serveDef endpoint. In this case
								we'll fallback to displaying plain text. We should be able to
								sanitize/render DocHTML on the front-end to make this consistent.
							*/}

							{!defObj.DocHTML && defObj.Docs && defObj.Docs.length &&
								<div className={base.mb3}>
									<div>{defDescrHidden && this.splitPlainDescr(defObj.Docs[0].Data, DESCRIPTION_CHAR_CUTOFF) || defObj.Docs[0].Data}</div>
									{defDescrHidden &&
										<a href="#" onClick={this._onViewMore} styleName="f7">View More...</a>
									}
									{!defDescrHidden && this.shouldHideDescr(defObj, DESCRIPTION_CHAR_CUTOFF) &&
										<a href="#" onClick={this._onViewLess} styleName="f7">Collapse</a>
									}
								</div>
							}

							{defObj.DocHTML &&
								<div>
									<div className={base.mb3}>
										<div dangerouslySetInnerHTML={defDescrHidden && {__html: this.splitHTMLDescr(defObj.DocHTML.__html, DESCRIPTION_CHAR_CUTOFF)} || defObj.DocHTML}></div>
										{defDescrHidden &&
											<a href="#" onClick={this._onViewMore} styleName="f7">View More...</a>
										}
										{!defDescrHidden && this.shouldHideDescr(defObj, DESCRIPTION_CHAR_CUTOFF) &&
											<a href="#" onClick={this._onViewLess} styleName="f7">Collapse</a>
										}
									</div>

								</div>
							}

							<div styleName="f7 cool_mid_gray">
								{defObj && defObj.Repo && <Link to={urlToRepo(defObj.Repo)} styleName="link_subtle" onClick={this._viewRepoClicked.bind(this)}>{defObj.Repo}</Link>}
								&nbsp; &middot; &nbsp;
								<Link title="View definition in code" to={defBlobUrl} onClick={this._viewDefinitionClicked.bind(this)} styleName="link_subtle">View definition</Link>
							</div>


							{!refLocations && <div className={typography.tc}><Loader /></div>}

							{refLocations &&
								<div>
									{refLocations.RepoRefs &&
										<div>
											{examples &&
												<div className={base.mt5}>
													<ExamplesContainer
														repo={repo}
														rev={rev}
														commitID={defCommitID}
														def={def}
														defObj={defObj}
														examples={examples} />
												</div>
											}
											<div className={base.mt5}>
												<RepoRefsContainer
													repo={repo}
													rev={rev}
													commitID={defCommitID}
													def={def}
													defObj={defObj}
													refLocations={refLocations} />
											</div>
										</div>
									}

									{!refLocations.RepoRefs &&
										<div className={`${typography.tc} ${base.center} ${base.mv5}`} style={{maxWidth: "500px"}}>
											<EmptyNodeIllo className={base.mv3} />
											<Heading level="5">
												We can't find any usage examples or <br className={base["hidden_s"]} />
												references for this definition
											</Heading>
											<p styleName="cool_mid_gray">
												It looks like this node in the graph is missing.
												{!(this.context as any).signedIn &&
													<span> Help us get more nodes in the graph by joining with GitHub.</span>
												}
											</p>
											{!(this.context as any).signedIn &&
												<p className={base.mt4}><GitHubAuthButton size="small">Join with GitHub</GitHubAuthButton></p>
											}
										</div>
									}

								</div>
							}
						</div>
					}
				</div>
			</FlexContainer>
		);
	}
}

export default CSSModules(DefInfo, styles, {allowMultiple: true});
