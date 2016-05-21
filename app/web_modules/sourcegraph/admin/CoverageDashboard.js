import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./CoverageBackend"; // for side effects
import CoverageStore from "sourcegraph/admin/CoverageStore";
import CoverageGraph from "sourcegraph/admin/CoverageGraph";
import * as CoverageActions from "sourcegraph/admin/CoverageActions";

import CSSModules from "react-css-modules";
import styles from "./styles/Coverage.css";

const langTargets = {
	"Go": 0.95,
	"JavaScript": 0.5,
	"C#": 0.5,
	"CSS": 0.75,
};

class CoverageDashbaord extends Container {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.coverage = CoverageStore.coverage;
		if (state.coverage && !state.coverage.Error && !state.processedCoverage) {
			// This computation may be fairly expensive; make sure we only do it once.

			let cvgByLangByDay = {}; // holds summation of all repo coverage data
			state.coverage.forEach((cvg) => {
				if (!cvg.Summary) return;
				cvg.Summary.forEach((summary) => {
					if (!cvgByLangByDay[summary.Language]) cvgByLangByDay[summary.Language] = {};
					if (!cvgByLangByDay[summary.Language][cvg.Day]) cvgByLangByDay[summary.Language][cvg.Day] = {Idents: 0, Refs: 0, Defs: 0};
					cvgByLangByDay[summary.Language][cvg.Day].Idents += summary.Idents;
					cvgByLangByDay[summary.Language][cvg.Day].Refs += summary.Refs;
					cvgByLangByDay[summary.Language][cvg.Day].Defs += summary.Defs;
				});
			});

			state.data = {};
			Object.keys(cvgByLangByDay).forEach((lang) => {
				const langData = Object.keys(cvgByLangByDay[lang]).map((day) => {
					const dayTotals = cvgByLangByDay[lang][day];
					return {Day: day, Refs: dayTotals.Refs / dayTotals.Idents, Defs: dayTotals.Defs / dayTotals.Idents};
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

	stores() { return [CoverageStore]; }

	render() {
		return (
			<div styleName="container">
				{this.state.data && Object.keys(this.state.data).map((lang, i) =>
					<div styleName="graph" key={i}>
						<div styleName="title">{lang}</div>
						<CoverageGraph data={this.state.data[lang]} target={langTargets[lang]} />
					</div>
				)}
			</div>
		);
	}
}

export default CSSModules(CoverageDashbaord, styles);
