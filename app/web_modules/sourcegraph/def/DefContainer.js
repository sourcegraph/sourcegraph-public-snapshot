// @flow weak

import React from "react";
import CSSModules from "react-css-modules";

import Blob from "sourcegraph/blob/Blob";
import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import {Link} from "react-router";
import "sourcegraph/blob/BlobBackend";
import {routeParams as defRouteParams} from "sourcegraph/def";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import styles from "./styles/Refs.css";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {FaAngleDown, FaAngleRight} from "sourcegraph/components/Icons";
import breadcrumb from "sourcegraph/util/breadcrumb";
import stripDomain from "sourcegraph/util/stripDomain";
import type {Def} from "sourcegraph/def";

class DefContainer extends Container {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		def: React.PropTypes.string.isRequired,
		defObj: React.PropTypes.object.isRequired,
	};

	stores() {
		return [DefStore, BlobStore];
	}

	state: {
		repo: string;
		showDef: boolean;
		mouseover: boolean;
		rev: ?string;
		commitID: ?string;
		def: string;
		defObj: Def;
	} = {
		repo: "",
		showDef: false,
		commitID: null,
		rev: null,
		mouseover: false,
		def: "",
		defObj: {DefStart: null, DefEnd: null},
	};

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.commitID = state.defObj && !state.defObj.Error ? state.defObj.CommitID : null;

		if (state.mouseover) {
			state.highlightedDef = DefStore.highlightedDef;
			if (state.highlightedDef) {
				let {repo, rev, def} = defRouteParams(state.highlightedDef);
				state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
			} else {
				state.highlightedDefObj = null;
			}
		}
		if (state.mouseout) {
			// Clear DefTooltip so it doesn't hang around.
			state.highlightedDef = null;
			state.highlightedDefObj = null;
		}

		state.defFile = state.defObj && !state.defObj.Error ? BlobStore.files.get(state.defObj.Repo, state.commitID, state.defObj.File) : null;
		state.defAnns = state.defObj && !state.defObj.Error ? BlobStore.annotations.get(state.defObj.Repo, state.commitID, state.defObj.File): null;
	}

	onStateTransition(prevState, nextState) {
		const defPropsUpdated = prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.def !== nextState.def || prevState.defObj !== nextState.defObj;
		const initialLoad = !prevState.repo && !prevState.commitID && !prevState.def && !prevState.defObj;
		if ((defPropsUpdated && !initialLoad) || (nextState.mouseover && !prevState.mouseover && defPropsUpdated)) {
			Dispatcher.Backends.dispatch(new BlobActions.WantFile(nextState.defObj.Repo, nextState.commitID, nextState.defObj.File));
			Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(nextState.defObj.Repo, nextState.commitID, nextState.defObj.File));
		}
	}

	renderFileHeader(def) {
		let path = stripDomain(def.Repo.concat("/", def.File));
		let pathBreadcrumb = breadcrumb(
			path,
			(j) => <span key={j} className={styles.sep}> / </span>,
			(_, component, j, isLast) => {
				let span = <span className={styles.pathPart} key={j}>{component}</span>;
				if (isLast) {
					return <Link className={styles.pathEnd} to={urlToBlob(def.Repo, this.state.rev, def.File)} key={j}> {span} </Link>;
				}
				return span;
			}
		);
		return (
			<div className={styles.filename} onClick={(e) => { this.setState({showDef: !this.state.showDef}); }}>
					<div className={styles.breadcrumbIcon}>
						{this.state.showDef ? <FaAngleDown className={styles.toggleIcon} /> : <FaAngleRight className={styles.toggleIcon} />}
					</div>
					<div className={styles.pathContainer}>
						{pathBreadcrumb}
					</div>
			</div>
		);
	}

	render() {
		let def = this.state.defObj;
		let deffile = def ? def.File : null;
		let beginningLine = this.state.defFile && !this.state.defFile.Error ? Math.max(lineFromByte(this.state.defFile.ContentsString, this.state.defObj.DefStart), 0) : null;
		// shows 15 lines of the def or the entire def, whichever is shorter
		let defRange = this.state.defFile && !this.state.defFile.Error ? [[
			beginningLine,
			Math.min(lineFromByte(this.state.defFile.ContentsString, this.state.defObj.DefEnd), beginningLine + 14),
		]] : null;
		let contents = this.state.defFile && !this.state.defFile.Error ? this.state.defFile.ContentsString : null;

		let errMsg;
		if (this.state.defFile && this.state.defFile.Error) {
			switch (this.state.defFile.Error.response.status) {
			case 413:
				errMsg = "Sorry, this file is too large to display.";
				break;
			default:
				errMsg = "File is not available.";
			}
		}

		return (
			<div className={styles.container}
				onMouseOver={() => this.setState({mouseover: true, mouseout: false})}
				onMouseOut={() => this.setState({mouseover: false, mouseout: true})}>
				{this.renderFileHeader(def)}
				{!this.state.showDef && this.state.defFile && !this.state.defFile.Error && beginningLine && <Blob
					repo={def.Repo}
					rev={this.state.rev}
					path={deffile}
					contents={contents}
					annotations={this.state.defAnns || null}
					activeDef={this.state.def}
					lineNumbers={true}
					displayRanges={[[beginningLine, beginningLine]]}
					highlightedDef={this.state.highlightedDef || null}
					highlightedDefObj={this.state.highlightedDefObj || null}
					displayLineExpanders="bottom"/>}
				{errMsg && <p className={styles.fileError}>{errMsg}</p>}
				{this.state.showDef && this.state.defFile && !this.state.defFile.Error && <Blob
					repo={def.Repo}
					rev={this.state.rev}
					path={deffile}
					contents={contents}
					annotations={this.state.defAnns || null}
					activeDef={this.state.def}
					lineNumbers={true}
					displayRanges={defRange || null}
					highlightedDef={this.state.highlightedDef || null}
					highlightedDefObj={this.state.highlightedDefObj || null} />}
				{this.state.highlightedDefObj && !this.state.highlightedDefObj.Error && <DefTooltip currentRepo={this.state.repo} def={this.state.highlightedDefObj} />}
			</div>
		);
	}
}

export default CSSModules(DefContainer, styles);
