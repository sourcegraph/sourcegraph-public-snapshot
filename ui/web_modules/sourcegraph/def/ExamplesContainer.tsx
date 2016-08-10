// tslint:disable: typedef ordered-imports curly

import * as React from "react";
import {Container} from "sourcegraph/Container";
import {RefsContainer} from "sourcegraph/def/RefsContainer";
import {DefStore} from "sourcegraph/def/DefStore";
import "sourcegraph/blob/BlobBackend";
import * as styles from "./styles/DefInfo.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import {Panel, Heading, Loader} from "sourcegraph/components/index";
import "whatwg-fetch";
import * as classNames from "classnames";

type Props = {
	repo: string,
	rev?: string,
	commitID?: string,
	def: string,
	defObj?: any,
	className?: string,
	examples?: any,
};

export class ExamplesContainer extends Container<Props, any> {
	constructor(props) {
		super(props);
	}

	stores() {
		return [DefStore];
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.defRepos = props.defRepos || [];
		state.sorting = props.sorting || null;
		state.examples = props.examples || null;
	}

	render(): JSX.Element | null {
		let refLocs = this.state.examples;

		const expandedSnippets = 3;
		return (
			<div>
				<Heading level="7" className={classNames(base.mb3, styles.cool_mid_gray)}>
					Usage Example{(refLocs && refLocs.RepoRefs && refLocs.RepoRefs.length > 1) ? "s" : ""}
				</Heading>
				<Panel
					hoverLevel="low"
					className={classNames(styles.full_width_sm, styles.b__cool_pale_gray, base.ba)}>
					<div className={this.props.className}>
						{!refLocs && <div className={typography.tc}> <Loader /></div>}
						{refLocs && !refLocs.RepoRefs && <i>No examples found</i>}
						{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.map((repoRefs, i) => <RefsContainer
							key={i}
							refIndex={i}
							repo={this.props.repo}
							rev={this.props.rev}
							def={this.props.def}
							defObj={this.props.defObj}
							repoRefs={repoRefs}
							initNumSnippets={expandedSnippets}
							rangeLimit={2}
							fileCollapseThreshold={5} />)}
					</div>
				</Panel>
			</div>
		);
	}
}
