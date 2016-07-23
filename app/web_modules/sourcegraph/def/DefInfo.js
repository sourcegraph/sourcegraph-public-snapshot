// @flow weak

import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";
import "sourcegraph/blob/BlobBackend";
import "whatwg-fetch";

import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import base from "sourcegraph/components/styles/_base.css";
import typography from "sourcegraph/components/styles/_typography.css";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import DefStore from "sourcegraph/def/DefStore";
import RepoRefsContainer from "sourcegraph/def/RepoRefsContainer";
import ExamplesContainer from "sourcegraph/def/ExamplesContainer";
import * as DefActions from "sourcegraph/def/DefActions";
import {urlToDef} from "sourcegraph/def/routes";
import {qualifiedNameAndType, defTitle, defTitleOK} from "sourcegraph/def/Formatter";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {trimRepo} from "sourcegraph/repo";
import {urlToRepo} from "sourcegraph/repo/routes";
import {EmptyNodeIllo} from "sourcegraph/components/symbols";
import {Header, Heading, FlexContainer, GitHubAuthButton, Loader} from "sourcegraph/components";

// Number of characters of the Docstring to show before showing the "collapse" options.
const DESCRIPTION_CHAR_CUTOFF = 80;
//
class DefInfo extends Container {
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
		return [DefStore];
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
			return doc.documentElement.textContent.length >= cutOff;
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
			let childrenCopy = [];
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
		this.setState({defDescrHidden: false});
	}

	_onViewLess() {
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
	}

	render() {
		const {defObj, defDescrHidden, refLocations, examples, repo, rev, defCommitID, def} = this.state;
		let defBlobUrl = defObj ? urlToDef(defObj, rev) : "";

		if (refLocations && refLocations.Error) {
			return (
				<Header
					title={`${httpStatusCode(refLocations.Error)}`}
					subtitle={`References are not available.`} />
			);
		}
		return (
			<FlexContainer styleName="bg-cool-pale-gray-2 flex-grow">
				<div styleName="container-fixed" className={base.mv3}>
					{/* NOTE: This should (roughly) be kept in sync with page titles in app/internal/ui. */}
					<Helmet title={defTitleOK(defObj) ? `${defTitle(defObj)} · ${trimRepo(repo)}` : trimRepo(repo)} />
					{defObj &&
						<div className={`${base.mv4} ${base.ph4}`}>
							<Heading level="5" styleName="break-word" className={base.mv2}>
								<code styleName="normal">{qualifiedNameAndType(defObj, {unqualifiedNameClass: styles.def})}</code>
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

							<div styleName="f7 cool-mid-gray">
								{defObj && defObj.Repo && <Link to={urlToRepo(defObj.Repo)} styleName="link-subtle">{defObj.Repo}</Link>}
								&nbsp; &middot; &nbsp;
								<Link title="View definition in code" to={defBlobUrl} styleName="link-subtle">View definition</Link>
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
												We can't find any usage examples or <br className={base["hidden-s"]} />
												references for this definition
											</Heading>
											<p styleName="cool-mid-gray">
												It looks like this node in the graph is missing.
												{!this.context.signedIn &&
													<span> Help us get more nodes in the graph by joining with GitHub.</span>
												}
											</p>
											{!this.context.signedIn &&
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
