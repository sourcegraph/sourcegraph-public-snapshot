// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as grid from "sourcegraph/components/styles/_grid.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import {Container} from "sourcegraph/Container";
import {UserStore} from "sourcegraph/user/UserStore";
import {Store} from "sourcegraph/Store";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {Heading, FlexContainer} from "sourcegraph/components";
import {locationForSearch} from "sourcegraph/search/routes";
import {GlobalSearchInput} from "sourcegraph/search/GlobalSearchInput";
import * as classNames from "classnames";
import {Link} from "react-router";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	location: any;
	currentStep?: string;
	completedBanner?: boolean;
}

type State = any;

type OnSelectQueryListener = (ev: React.MouseEvent<HTMLButtonElement>, query: string) => any;

const defaultSearchScope =  {popular: true, public: true, private: false, repo: false};

export class SignedInDashboard extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	_pageName: string = this.props.completedBanner ? "CompletedOnboardingDashboard" : "Dashboard";

	constructor(props) {
		super(props);
		this._handleChange = this._handleChange.bind(this);
	}

	reconcileState(state: State, props: Props, context: any): void {
		Object.assign(state, props);

		const settings = UserStore.settings;
		state.langs = settings && settings.search ? settings.search.languages : null;
	}

	_handleChange(ev: React.KeyboardEvent<HTMLInputElement>) {
		let value = (ev.currentTarget as HTMLInputElement).value;
		if (value) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_SUCCESS, "GlobalSearchInitiated", {page_name: this._pageName});
			this._goToSearch(value);
		}
	}

	_goToSearch(query: string) {
		(this.context as any).router.push(locationForSearch(this.props.location, query, this.state.langs, defaultSearchScope, true, true));
	}

	stores(): Store<any>[] {
		return [UserStore];
	}

	_topQuerySelected(query: string) {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "TopQuerySelected", {page_name: this._pageName, selected_query: query});
		this._goToSearch(query);
	}

	_exampleRepoSelected(exampleRepoUrl: string) {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ExampleRepoSelected", {page_name: this._pageName, example_repo: exampleRepoUrl});
	}

	_renderGlobalSearchForm(): JSX.Element | null {
		return (
			<div className={classNames(base.pt4, base.center)}>
				<GlobalSearchInput
					placeholder="Search for any function, symbol or package"
					name="q"
					showIcon={true}
					autoComplete="off"
					query={this.state.query || ""}
					onChange={this._handleChange} />
			</div>
		);
	}

	render(): JSX.Element | null {
		return (
			<div>
				<div className={styles.onboarding_container} style={{maxWidth: "750px"}}>
					<div className={classNames(base.pb3, base.ph4, base.br2)}>
						{this.props.completedBanner &&
							<div className={base.pt4}>
								<FlexContainer className={classNames(base.pv3, base.ph4, base.br2, colors.bg_green, base.center)}>
									<img src={`${(this.context as any).siteConfig.assetsRoot}/img/emoji/tada.svg`} style={{flex: "0 0 36px"}}/>
									<div className={base.pl3}>
										<h4 className={classNames(base.mv0, colors.white)}>Thanks for joining Sourcegraph!</h4>
										<span className={classNames(colors.white)}>Get started by searching for usage examples or exploring a public repository.</span>
									</div>
								</FlexContainer>
							</div>
						}
						<Heading className={classNames(base.pt5)} align="center" level="4">
							Start exploring code
						</Heading>
						<p className={classNames(typography.tc, base.mt3, base.mb4, typography.f6, colors.cool_gray_8)} >
							You've got everything you need to start exploring the code you depend on.
						</p>
						{this._renderGlobalSearchForm()}
						<div className={classNames(styles.user_actions, colors.cool_gray_8)}>
							Try these top searches:
							<a onClick={this._topQuerySelected.bind(this, "new http request")}> new http request</a>, <a onClick={this._topQuerySelected.bind(this, "read file")}>read file</a>, <a onClick={this._topQuerySelected.bind(this, "json encoder")}>json encoder</a>
						</div>
						<div className={classNames(styles.user_actions, base.pt5)}>
							<Heading className={base.pb4} level="5">Explore public repositories</Heading>
							<div style={{maxWidth: "675px"}} className={classNames(typography.tl, base.center, styles.repos_left_padding)}>
								<div className={classNames(colors.cool_gray_8, base.center)}>
									<div className={classNames(grid.col_6_ns, grid.col, base.pr5, base.pb3)}>
										<Link to="github.com/sourcegraph/checkup/-/blob/checkup.go"><span onClick={this._exampleRepoSelected.bind(this, "checkup")}>sourcegraph / checkup</span></Link>
										<p>Self-hosted health checks and status pages</p>
									</div>
									<div  className={classNames(grid.col_6_ns, grid.col, base.pr5, base.pb3)}>
										<Link to="github.com/gorilla/mux/-/blob/mux.go"><span onClick={this._exampleRepoSelected.bind(this, "mux")}>gorilla / mux</span></Link>
										<p>A powerful URL router and dispatcher for golang</p>
									</div>
								</div>
								<div className={classNames(colors.cool_gray_8, base.center)}>
									<div className={classNames(grid.col_6_ns, grid.col, base.pr5, base.pb3)}>
										<Link to="github.com/sourcegraph/thyme/-/blob/cmd/thyme/main.go"><span onClick={this._exampleRepoSelected.bind(this, "thyme")}>sourcegraph / thyme</span></Link>
										<p>Automatically track which applications you use</p>
									</div>
									<div  className={classNames(grid.col_6_ns, grid.col, base.pr5, base.pb3)}>
										<Link to="github.com/kubernetes/kubernetes/-/blob/examples/apiserver/server/main.go"><span onClick={this._exampleRepoSelected.bind(this, "kubernetes")}>kubernetes / kubernetes</span></Link>
										<p>Production-Grade Container Scheduling and Management</p>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}
