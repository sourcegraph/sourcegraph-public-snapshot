// @flow

import * as React from "react";
import {Link} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import {Avatar, Popover, Menu, Logo, Heading} from "sourcegraph/components";
import LogoutLink from "sourcegraph/user/LogoutLink";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalNav.css";
import base from "sourcegraph/components/styles/_base.css";
import colors from "sourcegraph/components/styles/_colors.css";
import typography from "sourcegraph/components/styles/_typography.css";

import {LoginForm} from "sourcegraph/user/Login";
import {EllipsisHorizontal, CheckIcon} from "sourcegraph/components/Icons";
import {DownPointer} from "sourcegraph/components/symbols";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import GlobalSearchInput from "sourcegraph/search/GlobalSearchInput";
import {locationForSearch, queryFromStateOrURL, langsFromStateOrURL, scopeFromStateOrURL} from "sourcegraph/search/routes";
import SearchResultsPanel from "sourcegraph/search/SearchResultsPanel";
import invariant from "invariant";
import {rel} from "sourcegraph/app/routePatterns";
import {repoPath, repoParam} from "sourcegraph/repo";
import {isPage} from "sourcegraph/page";
import debounce from "lodash.debounce";

function GlobalNav({navContext, location, params, channelStatusCode}, {user, siteConfig, signedIn, router, eventLogger}) {
	const isHomepage = location.pathname === "/";
	const isStaticPage = isPage(location.pathname);

	const showLogoMarkOnly = !isStaticPage || user;

	if (location.pathname === "/styleguide") return <span />;
	const repoSplat = repoParam(params.splat);
	let repo = repoSplat ? repoPath(repoSplat) : null;	return (
		<nav
			id="global-nav"
			styleName="navbar"
			className={colors["shadow-gray"]} role="navigation"
			style={isHomepage ? {visibility: "hidden"} : {}}>

			{location.state && location.state.modal === "login" &&
			// TODO: Decouple existence of modals and GlobalNav
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

			<div styleName="flex flex-fill flex-center tl" className={base.bn}>
				{!isHomepage && <Link to="/" styleName="logo-link flex-fixed">
					{showLogoMarkOnly ?
						<Logo styleName="logo logomark"
							width="21px"
							type="logomark"/> :
						<span>
							<Logo styleName="logo logomark small-only"
								width="21px"
								type="logomark"/>
							<Logo styleName="logo not-small-only"
								width="144px"
								type="logotype"/>
						</span>
					}
				</Link>}

				<div
					styleName="flex-fill"
					className={`${base["b--dotted"]} ${base.bn} ${base.brw2} ${colors["b--cool-pale-gray"]}`}>
					{location.pathname !== "/" && <SearchForm repo={repo} location={location} router={router} showResultsPanel={location.pathname !== `/${rel.search}`} />}
				</div>

				{typeof channelStatusCode !== "undefined" && channelStatusCode === 0 && <EllipsisHorizontal styleName="icon-ellipsis" title="Your editor could not identify the symbol"/>}
				{typeof channelStatusCode !== "undefined" && channelStatusCode === 1 && <CheckIcon styleName="icon-check" title="Sourcegraph successfully looked up symbol" />}

				{user && <div styleName="flex flex-fixed" className={`${base.pv2} ${base.ph3}`}>
					<Popover left={true}>
						<div styleName="user">
							{user.AvatarURL ? <Avatar size="small" img={user.AvatarURL} /> : <div>{user.Login}</div>}
							<DownPointer width={10} className={base.ml2} styleName="fill-cool-mid-gray" />
						</div>
						<Menu className={base.pa0} style={{width: "220px"}}>
							<div className={`${base.pa0} ${base.mb2} ${base.mt3}`}>
								<Heading level="7" color="cool-mid-gray">Signed in as</Heading>
								{user.Login}
							</div>
							<LogoutLink role="menu-item" />

							<hr role="divider" className={base.mv3} />
							<Link to="/settings/repos" role="menu-item">Your repositories</Link>
							<Link to="/tools" role="menu-item">Tools and integrations</Link>
							<hr role="divider" className={base.mt3} />
							<div styleName="cool-mid-gray" className={`${base.pv1} ${base.mb1} ${typography.tc}`}>
								<Link to="/security" className={`${typography.f7} ${typography["link-subtle"]} ${base.pr3}`} styleName="cool-mid-gray">Security</Link>
								<Link to="/-/privacy" className={`${typography.f7} ${typography["link-subtle"]} ${base.pr3}`} styleName="cool-mid-gray">Privacy</Link>
								<Link to="/-/terms" className={`${typography.f7} ${typography["link-subtle"]}`}>Terms</Link>
							</div>
						</Menu>
					</Popover>
				</div>}

				{!signedIn &&
					<div className={`${base.pv2} ${base.pr3} ${base.pl3}`}>
						<div>
							<LocationStateToggleLink href="/login" modalName="login" location={location}
								onToggle={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ShowLoginModal", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
								Sign in
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

// TODO(chexee): Move all these components to their own directory.

class SearchForm extends React.Component {
	static propTypes = {
		repo: React.PropTypes.string,
		location: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
		showResultsPanel: React.PropTypes.bool.isRequired,
	};

	constructor(props) {
		super(props);

		this.state.query = queryFromStateOrURL(props.location); // eslint-disable-line react/no-direct-mutation-state
		this.state.lang = langsFromStateOrURL(props.location); // eslint-disable-line react/no-direct-mutation-state
		this.state.scope = scopeFromStateOrURL(props.location); // eslint-disable-line react/no-direct-mutation-state

		this._handleGlobalHotkey = this._handleGlobalHotkey.bind(this);
		this._handleGlobalClick = this._handleGlobalClick.bind(this);
		this._handleSubmit = this._handleSubmit.bind(this);
		this._handleReset = this._handleReset.bind(this);
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._handleChange = this._handleChange.bind(this);
		this._handleFocus = this._handleFocus.bind(this);
		this._handleBlur = this._handleBlur.bind(this);
	}

	state: {
		open: bool;
		focused: bool;
		query: ?string;
		lang: ?string[];
		scope: ?Object;
	} = {
		open: false,
		focused: false,
		query: null,
		lang: null,
		scope: null,
	};

	componentDidMount() {
		document.addEventListener("keydown", this._handleGlobalHotkey);
		document.addEventListener("click", this._handleGlobalClick);
	}

	componentWillReceiveProps(nextProps) {
		const nextQuery = queryFromStateOrURL(nextProps.location);
		if (this.state.query !== nextQuery) {
			if (nextQuery && !this.state.query) {
				this.setState({open: true});
			} else {
				this.setState({query: nextQuery});
			}
		}

		if (!nextQuery) {
			this.setState({open: false});
		}
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
	_handleReset: any;
	_handleKeyDown: any;
	_handleChange: any;
	_handleFocus: any;
	_handleBlur: any;

	_handleGlobalHotkey(ev: KeyboardEvent) {
		if (ev.keyCode === 27 /* ESC */) {
			// Check that the element exists on the page before trying to set state.
			if (document.getElementById("e2etest-search-input")) {
				this.setState({
					open: false,
				});
			}
		}
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
		this.props.router.push(locationForSearch(this.props.location, this.state.query, this.state.lang, this.state.scope, false, true));
	}

	_handleReset(ev: Event) {
		this.props.router.push(locationForSearch(this.props.location, null, null, null, false, true));
		this.setState({focused: false, open: false});

		this.props.router.push(locationForSearch(this.props.location, this.state.query, this.state.lang, this.state.scope, true, true));
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
		const value = ev.currentTarget.value;
		this.setState({query: value});
		if (value) this.setState({open: true});
		this._goToDebounced(this.props.router.replace, locationForSearch(this.props.location, value, this.state.lang, this.state.scope, false, this.props.location.pathname.slice(1) === rel.search));
	}

	_goToDebounced = debounce((routerFunc: any, loc: Location) => {
		routerFunc(loc);
	}, 200, {leading: false, trailing: true});

	_handleFocus(ev: Event) {
		const update: {focused: boolean; open: boolean; query?: string} = {focused: true, open: true};
		if (this._input && this._input.value) {
			update.query = this._input.value;
		}
		this.setState(update);
	}

	_handleBlur(ev: Event) {
		this.setState({focused: false});
	}

	render() {
		if (this.state.open) {
			document.body.style.overflow = "hidden";
		} else if (!this.state.open) {
			document.body.style.overflow = "auto";
		}

		return (
			<div
				ref={e => this._container = e}>
				<form
					onSubmit={this._handleSubmit}
					styleName="flex"
					autoComplete="off">
					<GlobalSearchInput
						name="q"
						icon={true}
						autoComplete="off"
						query={this.state.query || ""}
						domRef={e => this._input = e}
						autoFocus={this.props.location.pathname.slice(1) === rel.search}
						onFocus={this._handleFocus}
						onBlur={this._handleBlur}
						onKeyDown={this._handleKeyDown}
						onClick={this._handleFocus}
						onChange={this._handleChange} />
						{this.props.showResultsPanel && this.state.open && <button styleName="close-icon" type="reset"></button>}
				</form>
				{this.props.showResultsPanel && this.state.open && <SearchResultsPanel query={this.state.query || ""} repo={this.props.repo} location={this.props.location} />}
			</div>
		);
	}
}
SearchForm = CSSModules(SearchForm, styles);
