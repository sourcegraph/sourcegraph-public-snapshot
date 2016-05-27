import React from "react";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Label} from "sourcegraph/components";
import {CloseIcon, TriangleLeftIcon, TriangleRightIcon} from "sourcegraph/components/Icons";

import BuildStore from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";
import {buildStatus, buildClass} from "sourcegraph/build/Build";
import {urlToBuilds} from "sourcegraph/build/routes";

import CSSModules from "react-css-modules";
import styles from "./styles/Coverage.css";

class CoverageDrilldown extends Container {
	static propTypes = {
		language: React.PropTypes.string.isRequired,
		data: React.PropTypes.arrayOf(React.PropTypes.shape({
			Day: React.PropTypes.string.isRequired,
			Refs: React.PropTypes.number.isRequired,
			Defs: React.PropTypes.number.isRequired,
			Sources: React.PropTypes.arrayOf(React.PropTypes.shape({
				Repo: React.PropTypes.string.isRequired,
				Language: React.PropTypes.string.isRequired,
				SrclibVersion: React.PropTypes.string,
				Summary: React.PropTypes.shape({
					Idents: React.PropTypes.number.isRequired,
					Refs: React.PropTypes.number.isRequired,
					Defs: React.PropTypes.number.isRequired,
				}),
			})).isRequired,
		})).isRequired,
		onDismiss: React.PropTypes.func.isRequired,
	}

	constructor(props) {
		super(props);
		this.state = {
			idx: props.data.length - 1, // most recent day
		};
		this.buildsQuery = "?Direction=desc&Ended=true&Sort=updated_at";
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

	refScore(summary) {
		if (summary.Idents === 0) return 0;
		return summary.Refs / summary.Idents;
	}

	defScore(summary) {
		if (summary.Idents === 0) return 0;
		return summary.Defs / summary.Idents;
	}

	formatDelta(delta) {
		if (delta) {
			return `${delta > 0 ? "+" : ""}${delta * 100}`.substring(0, 5);
		}
		return "";
	}

	refDelta(prevSummary, nextSummary) {
		return this.formatDelta(this.refScore(nextSummary) - this.refScore(prevSummary));
	}

	defDelta(prevSummary, nextSummary) {
		return this.formatDelta(this.defScore(nextSummary) - this.defScore(prevSummary));
	}

	deltaStyle(delta) {
		if (delta.indexOf("+") === 0) return "delta-increase";
		if (delta.indexOf("-") === 0) return "delta-decrease";
		return "";
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
										{source.Repo}
									</td>
									<td styleName="data">{summary ? summary.Idents : "---"}</td>
									<td styleName="data">
										{summary ? Math.round(this.refScore(summary) * 100) : "---"}
										<span styleName={this.deltaStyle(refDelta)}>{refDelta}</span>
									</td>
									<td styleName="data">
										{summary ? Math.round(this.defScore(summary) * 100) : "---"}
										<span styleName={this.deltaStyle(defDelta)}>{defDelta}</span>
									</td>
								</tr>
							);
						})}
					</tbody>
				</table>
			</div>
		);
	}
}

export default CSSModules(CoverageDrilldown, styles);
