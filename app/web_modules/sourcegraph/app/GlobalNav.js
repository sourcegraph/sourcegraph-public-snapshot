// @flow

import React from "react";
import {Link} from "react-router";
import type {RouterLocation} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import {Avatar, Panel, Popover, Menu, Button, TabItem, Logo} from "sourcegraph/components";
import LogoutLink from "sourcegraph/user/LogoutLink";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";
import base from "sourcegraph/components/styles/_base.css";
import {LoginForm} from "sourcegraph/user/Login";
import {EllipsisHorizontal, CheckIcon} from "sourcegraph/components/Icons";
import {FaChevronDown} from "sourcegraph/components/Icons";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import GlobalSearchInput from "sourcegraph/search/GlobalSearchInput";
import {locationForSearch, queryFromStateOrURL} from "sourcegraph/search/routes";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import SearchSettings from "sourcegraph/search/SearchSettings";
import invariant from "invariant";
import {rel} from "sourcegraph/app/routePatterns";
import {repoPath, repoParam} from "sourcegraph/repo";
import {isPage} from "sourcegraph/page";

function GlobalNav({navContext, location, params, channelStatusCode}, {user, siteConfig, signedIn, router, eventLogger}) {
	const isHomepage = location.pathname === "/";
	const isStaticPage = isPage(location.pathname);

	const showLogoMarkOnly = !isStaticPage || user;

	if (location.pathname === "/styleguide") return <span />;
	const repoSplat = repoParam(params.splat);
	let repo = repoSplat ? repoPath(repoSplat) : null;
	return (
		<nav id="global-nav" styleName={isHomepage ? "navbar-homepage" : "navbar-non-homepage"} role="navigation">

			{location.state && location.state.modal === "login" &&
				<LocationStateModal modalName="login" location={location}
					onDismiss={(v) => eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "DismissLoginModal", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
					<div styleName="modal">
						<LoginForm
							onLoginSuccess={dismissModal("login", location, router)}
							returnTo={location}
							location={location} />
					</div>
				</LocationStateModal>
			}

			<div styleName="flex flex-fill flex-center tl navbar-inner" className={base.bn}>
				{!isHomepage && <Link to="/" styleName="logo-link flex-fixed">
					{showLogoMarkOnly ?
						<Logo styleName={"logo logomark"}
							width={"21px"}
							type={"logomark"}/> :
						<span>
							<Logo styleName={"logo logomark small-only"}
								width={"21px"}
								type={"logomark"}/>
							<Logo styleName={"logo not-small-only"}
								width={"144px"}
								type={"logotype"}/>
						</span>
					}
				</Link>}

				<div styleName="search">
					{location.pathname !== "/" && <SearchForm repo={repo} location={location} router={router} showResultsPanel={location.pathname !== `/${rel.search}`} />}
				</div>

				{user && <div styleName="flex flex-start flex-fixed">
					<Link to="/settings/repos" styleName="nav-link">
						<TabItem active={location.pathname === "/settings/repos"}>Repositories</TabItem>
					</Link>
					<Link to="/tools" styleName="nav-link">
						<TabItem hideMobile={true} active={location.pathname === "/tools"}>Tools</TabItem>
					</Link>
				</div>}

				{typeof channelStatusCode !== "undefined" && channelStatusCode === 0 && <EllipsisHorizontal styleName="icon-ellipsis" title="Your editor could not identify the symbol"/>}
				{typeof channelStatusCode !== "undefined" && channelStatusCode === 1 && <CheckIcon styleName="icon-check" title="Sourcegraph successfully looked up symbol" />}

				{user && <div styleName="flex flex-fixed" className={`${base.pv2} ${base.ph3}`}>
					<Popover left={true}>
						<div styleName="user">
							{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} /> : <div>{user.Login}</div>}
							<FaChevronDown styleName="user-menu-icon" />
						</div>
						<Menu>
							<span styleName="current-user">Signed in as <strong>{user.Login}</strong></span>
							<hr className={base.m0} />
							<Link to="/about" role="menu-item">About</Link>
							<Link to="/contact" role="menu-item">Contact</Link>
							<Link to="/pricing" role="menu-item">Pricing</Link>
							<a href="https://text.sourcegraph.com" target="_blank" role="menu-item">Blog</a>
							<a href="https://boards.greenhouse.io/sourcegraph" target="_blank" role="menu-item">We're hiring</a>
							<Link to="/security" role="menu-item">Security</Link>
							<Link to="/-/privacy" role="menu-item">Privacy</Link>
							<Link to="/-/terms" role="menu-item">Terms</Link>
							<hr className={base.m0} />
							<LogoutLink role="menu-item" />
						</Menu>
					</Popover>
				</div>}

				{!signedIn &&
					<div styleName="tr" className={`${base.pv2} ${base.pr2}`}>
						<div styleName="action">
							<LocationStateToggleLink href="/login" modalName="login" location={location}
								onToggle={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ShowLoginModal", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
								<Button color="blue">Sign in</Button>
							</LocationStateToggleLink>
						</div>
					</div>
				}
			</div>
		</nav>
	);
}

GlobalNav.propTypes = {
	navContext: React.PropTypes.element,
	location: React.PropTypes.object.isRequired,
	params: React.PropTypes.object,
	channelStatusCode: React.PropTypes.number,
};
GlobalNav.contextTypes = {
	siteConfig: React.PropTypes.object.isRequired,
	user: React.PropTypes.object,
	signedIn: React.PropTypes.bool.isRequired,
	router: React.PropTypes.object.isRequired,
	eventLogger: React.PropTypes.object.isRequired,
};

export default CSSModules(GlobalNav, styles, {allowMultiple: true});

class SearchForm extends React.Component {
	// TODO(sqs): dismiss when click/focus outside

	static propTypes = {
		repo: React.PropTypes.string,
		location: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
		showResultsPanel: React.PropTypes.bool.isRequired,
	};

	constructor(props) {
		super(props);

		this.state.query = queryFromStateOrURL(props.location); // eslint-disable-line react/no-direct-mutation-state

		this._handleGlobalHotkey = this._handleGlobalHotkey.bind(this);
		this._handleGlobalClick = this._handleGlobalClick.bind(this);
		this._handleSubmit = this._handleSubmit.bind(this);
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._handleChange = this._handleChange.bind(this);
		this._handleFocus = this._handleFocus.bind(this);
		this._handleBlur = this._handleBlur.bind(this);
	}

	state: {
		open: bool;
		focused: bool;
		query: ?string;
	} = {
		open: false,
		focused: false,
		query: null,
	};

	componentDidMount() {
		document.addEventListener("keydown", this._handleGlobalHotkey);
		document.addEventListener("click", this._handleGlobalClick);
	}

	componentWillReceiveProps(nextProps) {
		const nextQuery = queryFromStateOrURL(nextProps.location);
		if (this.state.query !== nextQuery) {
			if (nextQuery && !this.state.query) this.setState({open: true});
			this.setState({query: nextQuery});
		}
		if (!nextQuery && !this.state.focused) this.setState({open: false});
	}

	componentWillUnmount() {
		document.removeEventListener("keydown", this._handleGlobalHotkey);
		document.removeEventListener("click", this._handleGlobalClick);
	}

	_container: HTMLElement;
	_input: HTMLInputElement;

	// NOTE: Flow doesn't automatically treat methods as props, so this manual list
	// is necessary. See https://github.com/facebook/flow/issues/1517.
	_handleGlobalHotkey: any;
	_handleGlobalClick: any;
	_handleSubmit: any;
	_handleKeyDown: any;
	_handleChange: any;
	_handleFocus: any;
	_handleBlur: any;

	_handleGlobalHotkey(ev: KeyboardEvent) {
		// Hotkey "/" to focus search field.
		invariant(this._input, "input not available");
		if (ev.keyCode === 191 /* forward slash "/" */) {
			if (!document.activeElement || (document.activeElement.tagName !== "INPUT" && document.activeElement.tagName !== "TEXTAREA" && document.activeElement.tagName !== "TEXTAREA")) {
				ev.preventDefault();
				this._input.focus();
			}
		}
	}

	_handleGlobalClick(ev: Event) {
		// Clicking outside of the open results panel should close it.
		invariant(ev.target instanceof Node, "target is not a node");
		if (this.state.open && (!this._container || !this._container.contains(ev.target))) {
			this.setState({open: false});
		}
	}

	_handleSubmit(ev: Event) {
		ev.preventDefault();
		this.props.router.push(locationForSearch(this.props.location, this.state.query, true, true));
	}

	_handleKeyDown(ev: KeyboardEvent) {
		if (ev.keyCode === 27 /* ESC */) {
			this.setState({open: false});
			this._input.blur();
		} else if (ev.keyCode === 13 /* Enter */) {
			// Close the search results menu AFTER the action has taken place on
			// the result (if a result was highlighted).
			setTimeout(() => this.setState({open: false}));
		}
	}

	_handleChange(ev: KeyboardEvent) {
		invariant(ev.currentTarget instanceof HTMLInputElement, "invalid currentTarget");
		this.props.router.replace(locationForSearch(this.props.location, ev.currentTarget.value, false, this.props.location.pathname.slice(1) === rel.search));
	}

	_handleFocus(ev: Event) {
		this.setState({focused: true, open: true});
	}

	_handleBlur(ev: Event) {
		this.setState({focused: false});
	}

	render() {
		return (
			<div
				styleName="search-form-container"
				ref={e => this._container = e}>
				<form
					onSubmit={this._handleSubmit}
					styleName="search-form"
					autoComplete="off">
					<GlobalSearchInput
						name="q"
						icon={true}
						autoComplete="off"
						styleName="search-input"
						query={this.state.query || ""}
						domRef={e => this._input = e}
						autoFocus={this.props.location.pathname.slice(1) === rel.search}
						onFocus={this._handleFocus}
						onBlur={this._handleBlur}
						onKeyDown={this._handleKeyDown}
						onClick={this._handleFocus}
						onChange={this._handleChange} />
				</form>
				{this.props.showResultsPanel && this.state.open && <SearchResultsPanel repo={this.props.repo} location={this.props.location} />}
			</div>
		);
	}
}
SearchForm = CSSModules(SearchForm, styles);

let SearchResultsPanel = ({repo, location}: {repo: ?string, location: RouterLocation}) => {
	const q = queryFromStateOrURL(location);
	return (
		<Panel hoverLevel="high" styleName="search-panel">
			<SearchSettings styleName="search-settings" innerClassName={styles["search-settings-inner"]} location={location} showAlerts={true} repo={repo} />
			<GlobalSearch styleName="search-results" query={q || ""} repo={repo} location={location} resultClassName={styles["search-result"]} />
		</Panel>
	);
};
SearchResultsPanel = CSSModules(SearchResultsPanel, styles);
