import * as React from "react";
import {InjectedRouter} from "react-router";

import {Location} from "history";
import * as invariant from "invariant";
import * as debounce from "lodash/debounce";

import {rel} from "sourcegraph/app/routePatterns";
import {langsFromStateOrURL, locationForSearch, queryFromStateOrURL, scopeFromStateOrURL} from "sourcegraph/search/routes";

import {EventListener} from "sourcegraph/Component";
import {Dismiss} from "sourcegraph/components/symbols/Dismiss";
import {GlobalSearchInput} from "sourcegraph/search/GlobalSearchInput";

import {colors} from "sourcegraph/components/utils";

import {SearchResultsPanel} from "sourcegraph/search/SearchResultsPanel";

interface Props {
	repo: string | null;
	location: Location;
	router: InjectedRouter;
	showResultsPanel: boolean;
	style: Object;
}

interface State {
	open: boolean;
	focused: boolean;
	query: string | null;
	lang: string[] | null;
	scope: any;
}

export class GlobalSearchContainer extends React.Component<Props, State> {

	_container: HTMLElement;
	_input: HTMLInputElement;

	_goToDebounced: any = debounce((routerFunc: any, loc: Location) => {
		routerFunc(loc);
	}, 200, {leading: false, trailing: true});

	state: State = {
		open: false,
		focused: false,
		query: null,
		lang: null,
		scope: null,
	};

	constructor(props: Props) {
		super(props);

		this.state.query = queryFromStateOrURL(props.location);
		this.state.lang = langsFromStateOrURL(props.location);
		this.state.scope = scopeFromStateOrURL(props.location);

		this._handleGlobalHotkey = this._handleGlobalHotkey.bind(this);
		this._handleGlobalClick = this._handleGlobalClick.bind(this);
		this._handleSubmit = this._handleSubmit.bind(this);
		this._handleReset = this._handleReset.bind(this);
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._handleChange = this._handleChange.bind(this);
		this._handleFocus = this._handleFocus.bind(this);
		this._handleBlur = this._handleBlur.bind(this);
	}

	componentWillReceiveProps(nextProps: Props): any {
		const nextQuery = queryFromStateOrURL(nextProps.location);
		if (this.state.query !== nextQuery) {
			if (nextQuery && !this.state.query) {
				this.setState({open: true} as State);
			} else {
				this.setState({query: nextQuery} as State);
			}
		}

		if (!nextQuery) {
			this.setState({open: false} as State);
		}
	}

	_handleGlobalHotkey(ev: KeyboardEvent): any {
		if (ev.keyCode === 27 /* ESC */) {
			// Check that the element exists on the page before trying to set state.
			if (document.getElementById("e2etest-search-input")) {
				this.setState({open: false} as State);
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

	_handleGlobalClick(ev: Event): any {
		// Clicking outside of the open results panel should close it.
		invariant(ev.target instanceof Node, "target is not a node");
		if (this.state.open && (!this._container || !this._container.contains(ev.target as Node))) {
			this.setState({open: false} as State);
		}
	}

	_handleSubmit(ev: React.FormEvent<HTMLFormElement>): any {
		ev.preventDefault();
		this.props.router.push(locationForSearch(this.props.location, this.state.query, this.state.lang, this.state.scope, false, true));
	}

	_handleReset(ev: React.MouseEvent<HTMLButtonElement>): any {
		this.setState({focused: false, open: false, query: ""} as State);
		this._input.value = "";
	}

	_handleKeyDown(ev: React.KeyboardEvent<HTMLInputElement>): any {
		if (ev.keyCode === 27 /* ESC */) {
			this.setState({open: false} as State);
			this._input.blur();
		} else if (ev.keyCode === 13 /* Enter */) {
			// Close the search results menu AFTER the action has taken place on
			// the result (if a result was highlighted).
			setTimeout(() => this.setState({open: false} as State));
		}
	}

	_handleChange(ev: React.KeyboardEvent<HTMLInputElement>): any {
		invariant(ev.currentTarget instanceof HTMLInputElement, "invalid currentTarget");
		const value = (ev.currentTarget as HTMLInputElement).value;
		if (value) {
			this.setState({query: value} as State);
			this.setState({open: true} as State);
		}
		this._goToDebounced(this.props.router.replace, locationForSearch(this.props.location, value, this.state.lang, this.state.scope, false, this.props.location.pathname.slice(1) === rel.search) as any);
	}

	_handleFocus(ev: React.FocusEvent<HTMLInputElement>): any {
		const update: {focused: boolean; open: boolean; query?: string} = {focused: true, open: true};
		if (this._input && this._input.value) {
			update.query = this._input.value;
		}
		this.setState(update as State);
	}

	_handleBlur(ev: React.FocusEvent<HTMLInputElement>): any {
		this.setState({focused: false} as State);
	}

	render(): JSX.Element | null {
		const containerSx = Object.assign({},
			{	position: "relative" },
			this.props.style,
		);

		const dismissBtnSx = {
			backgroundColor: "transparent",
			border: "0",
			position: "absolute",
			right: "6px",
			top: "4px",
		};

		return (
			<div
				ref={e => this._container = e}
				style={containerSx}>
				<form
					onSubmit={this._handleSubmit}
					autoComplete="off">
					<GlobalSearchInput
						name="q"
						showIcon={true}
						autoComplete="off"
						query={this.state.query || ""}
						domRef={e => this._input = e}
						autoFocus={this.props.location.pathname.slice(1) === rel.search}
						onFocus={this._handleFocus}
						onBlur={this._handleBlur}
						onKeyDown={this._handleKeyDown}
						onClick={this._handleFocus}
						onChange={this._handleChange} />
						{this.props.showResultsPanel && this.state.open &&
							<button style={dismissBtnSx} type="reset" onClick={this._handleReset}>
								<Dismiss color={colors.coolGray3()}/>
							</button>
						}
				</form>
				{this.props.showResultsPanel && this.state.open && <SearchResultsPanel query={this.state.query || ""} repo={this.props.repo} location={this.props.location} />}
				<EventListener target={global.document} event="keydown" callback={this._handleGlobalHotkey} />
				<EventListener target={global.document} event="click" callback={this._handleGlobalClick} />
			</div>
		);
	}
}
