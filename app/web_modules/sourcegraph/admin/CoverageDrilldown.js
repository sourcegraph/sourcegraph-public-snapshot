import React from "react";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Label, Modal} from "sourcegraph/components";
import {CloseIcon, TriangleLeftIcon, TriangleRightIcon, MagnifyingGlassIcon, FileIcon} from "sourcegraph/components/Icons";

import BuildStore from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";
import {buildStatus, buildClass} from "sourcegraph/build/Build";
import {urlToBuilds} from "sourcegraph/build/routes";
import {urlToRepoRev} from "sourcegraph/repo/routes";
import {urlToBlob} from "sourcegraph/blob/routes";

import CSSModules from "react-css-modules";
import styles from "./styles/Coverage.css";

class CoverageDrilldown extends Container {
	static propTypes = {
		language: React.PropTypes.string.isRequired,
		location: React.PropTypes.object.isRequired,
		refScore: React.PropTypes.func.isRequired,
		defScore: React.PropTypes.func.isRequired,
		data: React.PropTypes.arrayOf(React.PropTypes.shape({
			Day: React.PropTypes.string.isRequired,
			Refs: React.PropTypes.number.isRequired,
			Defs: React.PropTypes.number.isRequired,
			Sources: React.PropTypes.arrayOf(React.PropTypes.shape({
				Repo: React.PropTypes.string.isRequired,
				Rev: React.PropTypes.string,
				Language: React.PropTypes.string.isRequired,
				SrclibVersion: React.PropTypes.string,
				Summary: React.PropTypes.shape({
					Idents: React.PropTypes.number.isRequired,
					Refs: React.PropTypes.number.isRequired,
					Defs: React.PropTypes.number.isRequired,
				}),
				Files: React.PropTypes.arrayOf(React.PropTypes.shape({
					Path: React.PropTypes.string.isRequired,
					Idents: React.PropTypes.number.isRequired,
					Refs: React.PropTypes.number.isRequired,
					Defs: React.PropTypes.number.isRequired,
					UnresolvedIdents: React.PropTypes.arrayOf(React.PropTypes.shape({
						Offset: React.PropTypes.number.isRequired,
						Line: React.PropTypes.number.isRequired,
						Text: React.PropTypes.string.isRequired,
					})),
					UnresolvedRefs: React.PropTypes.arrayOf(React.PropTypes.shape({
						Offset: React.PropTypes.number.isRequired,
						Line: React.PropTypes.number.isRequired,
						Text: React.PropTypes.string.isRequired,
					})),
				})),
			})).isRequired,
		})).isRequired,
		onDismiss: React.PropTypes.func.isRequired,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			idx: props.data.length - 1, // most recent day
		};
		this.buildsQuery = "?Direction=desc&Ended=true&Sort=updated_at";
		this.getRepoDrilldown = this.getRepoDrilldown.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();

		// Instiantiate repo drilldown modal, if necessary.
		if (this.props.location.query.repo) {
			const repo = decodeURIComponent(this.props.location.query.repo);
			const lastDay = this.props.data[this.props.data.length - 1];
			for (const source of lastDay.Sources) {
				if (source.Repo === repo) {
					return this._drilldown(source);
				}
			}
		}
	}

	_drilldown(source) {
		let drilldownFiles;
		if (source) {
			// Sort files by increasing ref score.
			// To sort a deeploy frozen array, we must first create a copy of the array.
			drilldownFiles = source.Files.map((f) => Object.assign({}, f)).sort((a, b) => this.props.refScore(a) - this.props.refScore(b));
		}
		this.setState({drilldown: source || null, drilldownFiles: drilldownFiles || null}, () => {
			this.context.router.replace({...this.props.location, query: {lang: this.props.language, repo: source && source.Repo || undefined}}); // eslint-disable-line no-undefined
		});
	}

	stores() { return [BuildStore]; }

	reconcileState(state, props) {
		Object.assign(state, props);
		state.buildLists = BuildStore.buildLists;
	}

	onStateTransition(prevState, nextState) {
		if (!prevState.data && nextState.data) {
			let repos = {};
			nextState.data.forEach((datum) => {
				datum.Sources.forEach((source) => repos[source.Repo] = true);
			});
			Object.keys(repos).forEach((repo) => {
				Dispatcher.Backends.dispatch(new BuildActions.WantBuilds(repo, this.buildsQuery));
			});
		}
	}

	prevDay() {
		let idx = this.state.idx - 1;
		if (idx < 0) idx = this.props.data.length - 1;
		this.setState({idx: idx});
	}

	nextDay() {
		let idx = this.state.idx + 1;
		if (idx >= this.props.data.length) idx = 0;
		this.setState({idx: idx});
	}

	formatScore(score) {
		return Math.round(score * 100);
	}

	formatDelta(delta) {
		if (delta) {
			return `${delta > 0 ? "+" : ""}${delta * 100}`.substring(0, 5);
		}
		return "";
	}

	refDelta(prevSummary, nextSummary) {
		return this.formatDelta(this.props.refScore(nextSummary) - this.props.refScore(prevSummary));
	}

	defDelta(prevSummary, nextSummary) {
		return this.formatDelta(this.props.defScore(nextSummary) - this.props.defScore(prevSummary));
	}

	deltaStyle(delta) {
		if (delta.indexOf("+") === 0) return "delta-increase";
		if (delta.indexOf("-") === 0) return "delta-decrease";
		return "";
	}

	getRepoDrilldown(files) {
		return this.state.drilldownFiles.map((file, i) => {
			const blobURL = urlToBlob(this.state.drilldown.Repo, this.state.drilldown.Rev, file.Path);
			return (<div key={i}>
				<div styleName="file-drilldown-row">
					<div styleName={`file-drilldown-header${this.props.refScore(file) <= 0.5 ? "-uncovered" : ""}`} onClick={() => {
						if (this.state.drilldownFile === i) return this.setState({drilldownFile: null});
						this.setState({drilldownFile: i});
					}}>
						<div styleName="filepath">{file.Path}</div>
						<div styleName="file-stats">{`Idents (${file.Idents}) Refs (${this.formatScore(this.props.refScore(file))}%) Defs (${this.formatScore(this.props.defScore(file))}%)`}</div>
					</div>
					<Link styleName="file-link" to={blobURL}>
						<FileIcon />
					</Link>
				</div>
				{this.state.drilldownFile === i && <div styleName="unresolved-tokens">
					<div styleName="unresolved-idents">
						<div styleName="unresolved-header">Unresolved idents</div>
						{file.UnresolvedIdents && file.UnresolvedIdents.map((ident, j) => <div key={j}>
							<Link to={`${blobURL}#L${ident.Line}`}>{ident.Line}</Link>
							{` : ${ident.Text}`}
						</div>)}
					</div>
					<div styleName="unresolved-refs">
						<div styleName="unresolved-header">Unresolved refs</div>
						{file.UnresolvedRefs && file.UnresolvedRefs.map((ref, j) => <div key={j}>
							<Link to={`${blobURL}#L${ref.Line}`}>{ref.Line}</Link>
							{` : ${ref.Text}`}
						</div>)}
					</div>
				</div>}
			</div>);
		});
	}

	render() {
		const datum = this.props.data[this.state.idx];
		let srclibVersions = {};
		datum.Sources.forEach((source) => {
			if (source.SrclibVersion) srclibVersions[source.SrclibVersion] = true;
		});

		const prevDatum = this.state.idx === 0 ? null : this.props.data[this.state.idx - 1];
		let prevSummaryIndex = {};
		if (prevDatum) {
			prevDatum.Sources.forEach((source) => {
				prevSummaryIndex[source.Repo] = source.Summary;
			});
		}

		return (
			<div styleName="drilldown">
				<h2 styleName="drilldown-header">
					{this.props.language}
					{Object.keys(srclibVersions).map((ver, i) =>
						<span key={i} styleName="srclib-version">{ver}</span>
					)}
				</h2>
				<div>
					<Button styleName="day-chooser" size="small" outline={true} onClick={this.prevDay.bind(this)}><TriangleLeftIcon /></Button>
					<span styleName="day">{datum.Day}</span>
					<Button styleName="day-chooser" size="small" outline={true} onClick={this.nextDay.bind(this)}><TriangleRightIcon /></Button>
					<Button styleName="drilldown-dismiss" size="small" outline={true} onClick={this.props.onDismiss}><CloseIcon /></Button>
				</div>
				<table styleName="table">
					<thead>
						<tr>
							<th styleName="repo">Repo</th>
							<th styleName="idents">Idents</th>
							<th styleName="refs">Refs (%)</th>
							<th styleName="defs">Defs (%)</th>
						</tr>
					</thead>
					<tbody>
						{datum.Sources.map((source, i) => {
							const summary = source.Summary;
							const builds = this.state.buildLists.get(source.Repo, this.buildsQuery);
							const prevSummary = prevSummaryIndex[source.Repo];
							const refDelta = prevSummary && summary ? this.refDelta(prevSummary, summary) : "";
							const defDelta = prevSummary && summary ? this.defDelta(prevSummary, summary) : "";
							return (
								<tr key={i}>
									<td styleName="data">
										{builds && builds.length > 0 && <Link to={urlToBuilds(source.Repo)}>
											<Label color={buildClass(builds[0])} styleName="build-label">{buildStatus(builds[0])}</Label>
											</Link>
										}
										<Link to={urlToRepoRev(source.Repo, source.Rev)}>{source.Repo}</Link>
										{this.state.idx === this.props.data.length - 1 &&
											<div styleName="repo-drilldown-icon" size="small" outline={true} onClick={() => this._drilldown(source)}><MagnifyingGlassIcon /></div>
										}
									</td>
									<td styleName="data">{summary ? summary.Idents : "---"}</td>
									<td styleName="data">
										{summary ? this.formatScore(this.props.refScore(summary)) : "---"}
										<span styleName={this.deltaStyle(refDelta)}>{refDelta}</span>
									</td>
									<td styleName="data">
										{summary ? this.formatScore(this.props.defScore(summary)) : "---"}
										<span styleName={this.deltaStyle(defDelta)}>{defDelta}</span>
									</td>
								</tr>
							);
						})}
					</tbody>
				</table>
				{this.state.drilldown && <Modal onDismiss={() => this._drilldown(null)}>
					<div styleName="repo-drilldown-modal">
						<h3>
							<Link to={urlToRepoRev(this.state.drilldown.Repo, this.state.drilldown.Rev)}>
								{this.state.drilldown.Repo}@{this.state.drilldown.Rev}
							</Link>
						</h3>
						{this.getRepoDrilldown()}
					</div>
				</Modal>}
			</div>
		);
	}
}

export default CSSModules(CoverageDrilldown, styles);
