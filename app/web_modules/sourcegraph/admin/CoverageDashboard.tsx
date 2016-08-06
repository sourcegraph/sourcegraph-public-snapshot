// tslint:disable

import * as React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./CoverageBackend"; // for side effects
import CoverageStore from "sourcegraph/admin/CoverageStore";
import CoverageGraph from "sourcegraph/admin/CoverageGraph";
import CoverageDrilldown from "sourcegraph/admin/CoverageDrilldown";
import * as CoverageActions from "sourcegraph/admin/CoverageActions";
import {MagnifyingGlassIcon} from "sourcegraph/components/Icons";

import CSSModules from "react-css-modules";
import * as styles from "./styles/Coverage.css";

const langTargets = {
	"Go": 0.95,
	"JavaScript": 0.5,
	"C#": 0.5,
	"CSS": 0.75,
};

class CoverageDashboard extends Container<any, any> {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
	}

	_drilldown(lang) {
		(this.context as any).router.replace(Object.assign({}, this.props.location, {query: {lang: lang || undefined}})); // eslint-disable-line no-undefined
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.drilldown = props.location.query.lang || null;

		state.coverage = CoverageStore.coverage;
		if (state.coverage && !state.coverage.Error && !state.processedCoverage) {
			let cvgByLangByDay = {};
			state.coverage.forEach((cvg) => {
				if (!cvgByLangByDay[cvg.Language]) cvgByLangByDay[cvg.Language] = {};
				if (!cvgByLangByDay[cvg.Language][cvg.Day]) cvgByLangByDay[cvg.Language][cvg.Day] = {RefScores: [], DefScores: [], Sources: []};
				cvgByLangByDay[cvg.Language][cvg.Day].Sources.push(cvg);
				if (cvg.Summary) {
					cvgByLangByDay[cvg.Language][cvg.Day].RefScores.push(this.refScore(cvg.Summary));
					cvgByLangByDay[cvg.Language][cvg.Day].DefScores.push(this.defScore(cvg.Summary));
				}
			});

			state.data = {};
			Object.keys(cvgByLangByDay).forEach((lang) => {
				const langData = Object.keys(cvgByLangByDay[lang]).map((day) => {
					const dayObj = cvgByLangByDay[lang][day];
					dayObj.RefScores.sort((a, b) => a - b);
					dayObj.DefScores.sort((a, b) => a - b);

					const nRefScores = dayObj.RefScores.length;
					const nDefScores = dayObj.RefScores.length;

					const refAvg = nRefScores === 0 ? 0 : dayObj.RefScores.reduce((memo, val) => memo + val, 0) / nRefScores;
					const refQuantiles = [
						dayObj.RefScores[Math.floor(nRefScores / 4)],
						dayObj.RefScores[Math.floor(nRefScores / 2)],
						dayObj.RefScores[Math.floor(nRefScores * 3 / 4)],
					];

					const defAvg = nDefScores === 0 ? 0 : dayObj.DefScores.reduce((memo, val) => memo + val, 0) / nDefScores;
					const defQuantiles = [
						dayObj.DefScores[Math.floor(nDefScores / 4)],
						dayObj.DefScores[Math.floor(nDefScores / 2)],
						dayObj.DefScores[Math.floor(nDefScores * 3 / 4)],
					];

					return {Day: day, Refs: refAvg, RefQs: refQuantiles, Defs: defAvg, DefQs: defQuantiles, Sources: dayObj.Sources};
				});
				state.data[lang] = langData.sort((a, b) => {
					if (a.Day < b.Day) return -1;
					if (a.Day > b.Day) return 1;
					return 0;
				});
			});

			state.processedCoverage = true;
		}
	}

	onStateTransition(prevState, nextState) {
		if (!nextState.coverage && nextState.coverage !== prevState.coverage) {
			Dispatcher.Backends.dispatch(new CoverageActions.WantCoverage());
		}
	}

	refScore(summary) {
		if (summary.Idents === 0) return 0;
		return summary.Refs / summary.Idents;
	}

	defScore(summary) {
		if (summary.Idents === 0) return 0;
		return summary.Defs / summary.Idents;
	}

	stores() { return [CoverageStore]; }

	render(): JSX.Element | null {
		return (
			<div styleName="container">
				{this.state.data && !this.state.drilldown && Object.keys(this.state.data).map((lang, i) => {
					const data = this.state.data[lang];
					return (<div styleName="graph" key={i}>
						<div styleName="title" onClick={() => this._drilldown(lang)}>
							{lang}
							<MagnifyingGlassIcon styleName="icon" />
						</div>
						<div styleName="quantiles">
							<span styleName="quantile_header">Ref Quantiles: </span>
							{/* show quantile data for most recent day only */}
							{data[data.length - 1].RefQs.map((q, j) =>
								<span styleName="quantile" key={j}>{`${Math.round(q * 100)}% (p=${0.25 * (j+1)})`}</span>)}
						</div>
						<div styleName="quantiles">
							<span styleName="quantile_header">Def Quantiles: </span>
							{/* show quantile data for most recent day only */}
							{data[data.length - 1].DefQs.map((q, j) =>
								<span styleName="quantile" key={j}>{`${Math.round(q * 100)}% (p=${0.25 * (j+1)})`}</span>)}
						</div>
						<CoverageGraph data={this.state.data[lang]} target={langTargets[lang]} />
					</div>);
				})}
				{this.state.data && this.state.drilldown &&
					<CoverageDrilldown
						data={this.state.data[this.state.drilldown]}
						refScore={this.refScore}
						defScore={this.defScore}
						location={this.state.location}
						language={this.state.drilldown}
						onDismiss={() => this._drilldown(null)} />}
			</div>
		);
	}
}

export default CSSModules(CoverageDashboard, styles);
