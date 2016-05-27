import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./CoverageBackend"; // for side effects
import CoverageStore from "sourcegraph/admin/CoverageStore";
import CoverageGraph from "sourcegraph/admin/CoverageGraph";
import CoverageDrilldown from "sourcegraph/admin/CoverageDrilldown";
import * as CoverageActions from "sourcegraph/admin/CoverageActions";
import {MagnifyingGlassIcon} from "sourcegraph/components/Icons";

import CSSModules from "react-css-modules";
import styles from "./styles/Coverage.css";

const langTargets = {
	"Go": 0.95,
	"JavaScript": 0.5,
	"C#": 0.5,
	"CSS": 0.75,
};

class CoverageDashbaord extends Container {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
	}

	_drilldown(lang) {
		this.context.router.replace({...this.props.location, query: {lang: lang || undefined}}); // eslint-disable-line no-undefined
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.drilldown = props.location.query.lang || null;

		state.coverage = CoverageStore.coverage;
		if (state.coverage && !state.coverage.Error && !state.processedCoverage) {
			let cvgByLangByDay = {};
			state.coverage.forEach((cvg) => {
				if (!cvgByLangByDay[cvg.Language]) cvgByLangByDay[cvg.Language] = {};
				if (!cvgByLangByDay[cvg.Language][cvg.Day]) cvgByLangByDay[cvg.Language][cvg.Day] = {Idents: 0, Refs: 0, Defs: 0, Sources: []};
				cvgByLangByDay[cvg.Language][cvg.Day].Sources.push(cvg);
				if (cvg.Summary) {
					cvgByLangByDay[cvg.Language][cvg.Day].Idents += cvg.Summary.Idents;
					cvgByLangByDay[cvg.Language][cvg.Day].Refs += cvg.Summary.Refs;
					cvgByLangByDay[cvg.Language][cvg.Day].Defs += cvg.Summary.Defs;
				}
			});

			state.data = {};
			Object.keys(cvgByLangByDay).forEach((lang) => {
				const langData = Object.keys(cvgByLangByDay[lang]).map((day) => {
					const dayObj = cvgByLangByDay[lang][day];
					const refScore = dayObj.Idents === 0 ? 0 : dayObj.Refs / dayObj.Idents;
					const defScore = dayObj.Idents === 0 ? 0 : dayObj.Defs / dayObj.Idents;
					return {Day: day, Refs: refScore, Defs: defScore, Sources: dayObj.Sources};
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
				{this.state.data && !this.state.drilldown && Object.keys(this.state.data).map((lang, i) =>
					<div styleName="graph" key={i}>
						<div styleName="title" onClick={() => this._drilldown(lang)}>
							{lang}
							<MagnifyingGlassIcon styleName="icon" />
						</div>
						<CoverageGraph data={this.state.data[lang]} target={langTargets[lang]} />
					</div>
				)}
				{this.state.data && this.state.drilldown &&
					<CoverageDrilldown
						data={this.state.data[this.state.drilldown]}
						language={this.state.drilldown}
						onDismiss={() => this._drilldown(null)} />}
			</div>
		);
	}
}

export default CSSModules(CoverageDashbaord, styles);
