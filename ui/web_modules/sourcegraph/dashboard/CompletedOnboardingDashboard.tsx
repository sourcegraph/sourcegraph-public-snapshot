// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import {Container} from "sourcegraph/Container";
import {UserStore} from "sourcegraph/user/UserStore";
import {Store} from "sourcegraph/Store";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {Heading, Table, Panel} from "sourcegraph/components/index";
import {locationForSearch} from "sourcegraph/search/routes";
import {GlobalSearchInput} from "sourcegraph/search/GlobalSearchInput";
import * as classNames from "classnames";
import {Link} from "react-router";

interface Props {
	location?: any;
	currentStep?: string;
}

type State = any;

type OnSelectQueryListener = (ev: React.MouseEvent<HTMLButtonElement>, query: string) => any;

const defaultSearchScope =  {popular: true, public: true, private: false, repo: false};

export class CompletedOnboardingDashboard extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

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
		if (!(ev.currentTarget instanceof HTMLInputElement)) {
			return;
		}
		if (ev.currentTarget.value) {
			(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_SUCCESS, "OnboardingGlobalSearchInitiated", {page_name: "CompletedOnboardingDashboard"});
			this._goToSearch(ev.currentTarget.value);
		}
	}

	_goToSearch(query: string) {
		(this.context as any).router.push(locationForSearch(this.props.location, query, this.state.langs, defaultSearchScope, true, true));
	}

	stores(): Store<any>[] {
		return [UserStore];
	}

	_topQuerySelected(query: string) {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "TopQuerySelected", {page_name: "CompletedOnboardingDashboard", selected_query: query});
		this._goToSearch(query);
	}

	_exampleRepoSelected(exampleRepoUrl: string) {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ExampleRepoSelected", {page_name: "CompletedOnboardingDashboard", example_repo: exampleRepoUrl});
	}

	_renderGlobalSearchForm(): JSX.Element | null {
		return (
			<div className={classNames(base.pl3, base.pt4)} style={{maxWidth: "550px", margin: "0 auto"}}>
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
				<div className={styles.onboarding_container}>
					<Panel className={classNames(base.pb3, base.ph4, base.ba, base.br2, colors.b__cool_pale_gray)}>
						<div className={base.pt4}>
							<div className={classNames(base.pb3, base.ph4, base.br2, colors.bg_green, base.hidden_s)} style={{maxWidth: "550px", margin: "0 auto"}}>
								<img width={35} style={{marginTop: "22px", float: "left", display: "inline", marginLeft: "-10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Dashboard/PartyPopper.png`}></img>
								<h4 className={classNames(base.mv0, colors.white)} style={{paddingLeft: "40px", paddingTop: "18px"}}>Thanks for joining Sourcegraph!</h4>
								<span className={classNames(base.mv0, base.pl3, colors.white)}>Get started by searching for some code or exploring a repository.</span>
							</div>
						</div>
						<Heading className={classNames(base.pt5)} align="center" level="">
							Start exploring code
						</Heading>
						<div className={classNames(styles.user_actions, base.pt2)} style={{maxWidth: "500px"}}>
							<p className={classNames(typography.tc, base.mt3, base.mb2, typography.f6, colors.cool_gray_8)} >
								You've got everything you need to start browsing code smarter. Get started by searching for usage examples or exploring repositories.
							</p>
						</div>
						{this._renderGlobalSearchForm()}
						<div className={classNames(styles.user_actions, colors.cool_gray_8)}>
							Try these top searches:
							<a onClick={this._topQuerySelected.bind(this, "new http request")}> new http request</a>, <a onClick={this._topQuerySelected.bind(this, "read file")}>read file</a>, <a onClick={this._topQuerySelected.bind(this, "json encoder")}>json encoder</a>
						</div>
						<div className={classNames(styles.user_actions, base.pt3, base.hidden_s)}>
							<h3 className={base.pb3}>Explore repositories</h3>
							<Table style={{width: "575px", paddingLeft: "60px", margin: "0 auto"}} className={classNames(typography.tl)}>
								<tbody>
									<tr className={classNames(base.pt3)}>
										<td className={base.pr5}><Link to="github.com/sourcegraph/checkup"><span onClick={this._exampleRepoSelected.bind(this, "checkup")}>sourcegraph / checkup</span></Link></td>
										<td className={base.pr5}><Link to="github.com/gorilla/mux"><span onClick={this._exampleRepoSelected.bind(this, "mux")}>gorilla / mux</span></Link></td>
									</tr>
									<tr className={classNames(colors.cool_gray_8)}>
										<td className={classNames(base.pb4, base.pr4)} style={{width: "200px"}}>Self-hosted health checks and status pages</td>
										<td className={classNames(base.pb4, base.pr4)} style={{width: "200px"}}>A powerful URL router and dispatcher for golang</td>
									</tr>
								</tbody>
								<tbody>
									<tr>
										<td><Link to="github.com/sourcegraph/thyme"><span onClick={this._exampleRepoSelected.bind(this, "thyme")}>sourcegraph / thyme</span></Link></td>
										<td><Link to="github.com/kubernetes/kubernetes"><span onClick={this._exampleRepoSelected.bind(this, "kubernetes")}>kubernetes / kubernetes</span></Link></td>
									</tr>
									<tr className={colors.cool_gray_8}>
										<td className={classNames(base.pb4, base.pr4)} style={{width: "200px"}}>Automatically track which applications you use</td>
										<td className={classNames(base.pb4, base.pr4)} style={{width: "200px"}}>Production-Grade Container Scheduling and Management</td>
									</tr>
								</tbody>
							</Table>
						</div>
					</Panel>
				</div>
			</div>
		);
	}
}
