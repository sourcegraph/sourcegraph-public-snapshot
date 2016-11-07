import fuzzysearch from "fuzzysearch";
import * as debounce from "lodash/debounce";
import * as React from "react";
import {Link} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {Component, EventListener} from "sourcegraph/Component";
import {Base, Heading, Input, Menu} from "sourcegraph/components";
import {Check, DownMenu} from "sourcegraph/components/symbols";
import {colors, typography, whitespace} from "sourcegraph/components/utils";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend";
import {urlWithRev} from "sourcegraph/repo/routes";
import * as styles from "sourcegraph/repo/styles/RevSwitcher.css";

interface Props {
	repo: string;
	rev: string;
	commitID: string;
	repoObj?: any;
	isCloning: boolean;

	// branches is RepoStore.branches.
	branches: any;

	// tags is RepoStore.tags.
	tags: any;

	// to construct URLs
	routes: any[];
	routeParams: RouteParams;
}

interface State extends Props {
	open?: boolean;
	query?: any;
	effectiveRev?: string;
}

export class RevSwitcher extends Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	_input: any;
	_debouncedSetQuery: any;
	_wrapper: any;

	constructor(props: Props) {
		super(props);
		this.state = {
			open: false,

			repo: props.repo,
			rev: props.rev,
			commitID: props.commitID,
			isCloning: props.isCloning,
			branches: props.branches,
			tags: props.tags,
			routes: props.routes,
			routeParams: props.routeParams,
		};
		this._closeDropdown = this._closeDropdown.bind(this);
		this._onToggleDropdown = this._onToggleDropdown.bind(this);
		this._onChangeQuery = this._onChangeQuery.bind(this);
		this._onClickOutside = this._onClickOutside.bind(this);
		this._onKeydown = this._onKeydown.bind(this);
		this._debouncedSetQuery = debounce((query) => {
			this.setState(Object.assign({}, this.state, { query: query }));
		}, 150, {leading: true, trailing: true});
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);

		// effectiveRev is the rev from the URL, or else the repo's default branch.
		state.effectiveRev = props.rev || (props.repoObj && !props.repoObj.Error ? props.repoObj.DefaultBranch : null);
	}

	onStateTransition(prevState: State, nextState: State): void {
		const becameOpen = nextState.open && nextState.open !== prevState.open;
		if (becameOpen || nextState.repo !== nextState.repo) {
			// Don't load when page loads until we become open.
			const initialLoad = !prevState.repo && !nextState.open;
			if (!initialLoad) {
				Dispatcher.Backends.dispatch(new RepoActions.WantBranches(nextState.repo));
				Dispatcher.Backends.dispatch(new RepoActions.WantTags(nextState.repo));
			}
		}
	}

	// abbrevRev shortens rev if it is an absolute commit ID.
	_abbrevRev(rev: string): string {
		return rev.length === 40 ? rev.substring(0, 12) : rev;
	}

	_loadingItem(itemType: string): JSX.Element {
		return <li className={styles.disabled}>Loading {itemType}&hellip;</li>;
	}

	_errorItem(): JSX.Element {
		return <li className={styles.disabled}>Error</li>;
	}

	_emptyItem(): JSX.Element {
		return <li className={styles.disabled}>None found</li>;
	}

	_item(name: string, commitID: string): JSX.Element {
		let isCurrent = name === this.state.effectiveRev;

		return (
			<div key={`r${name}.${commitID}`} role="menu_item">
				<Link
					to={this._revSwitcherURL(name)}
					title={commitID}
					onClick={this._closeDropdown}>
					<span style={{ display: "inline-block", width: 24 }}>
						{isCurrent && <Check width={16} style={{ fill: colors.coolGray3() }} />}
					</span>

					{name && <span style={{fontWeight: isCurrent ? "bold" : "normal"}}>
						{this._abbrevRev(name)}
					</span>}

				</Link>
			</div>
		);
	}

	_closeDropdown(): void {
		// HACK: If the user clicks to a rev that they have already loaded all
		// of the data for, the transition occurs synchronously and the dropdown
		// does not close for some reason. Bypassing this.setState and setting it
		// directly fixes this issue.
		this.state.open = false;
		this.setState(Object.assign({}, this.state, { open: false }));
	}

	// If path is not present, it means this is the rev switcher on commits page.
	_revSwitcherURL(rev: string | null): string {
		return `${urlWithRev(this.state.routes, this.state.routeParams, rev)}${window.location.hash}`;
	}

	_onToggleDropdown(ev: React.MouseEvent<HTMLElement>): void {
		ev.preventDefault();
		ev.stopPropagation();
		this.setState(Object.assign({}, this.state, { open: !this.state.open }), () => {
			if (this.state.open && this._input) {
				this._input.focus();
			}
		});
	}

	_onChangeQuery(): void {
		if (this._input) {
			this._debouncedSetQuery(this._input.value);
		}
	}

	// _onClickOutside causes clicks outside the menu to close the menu.
	_onClickOutside(ev: MouseEvent): void {
		if (!this.state.open) {
			return;
		}
		if (this._wrapper && !this._wrapper.contains(ev.target)) {
			this.setState(Object.assign({}, this.state, { open: false }));
		}
	}

	// _onKeydown causes ESC to close the menu.
	_onKeydown(ev: KeyboardEvent): void {
		if (ev.defaultPrevented) {
			return;
		}

		// Don't trigger if there's a modifier key or if the cursor is focused
		// in an input field.
		const el = ev.target as HTMLElement;
		const tag = el.tagName;
		if (!(ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey) && typeof document !== "undefined" && tag !== "INPUT" && tag !== "TEXTAREA" && tag !== "SELECT") {
			// Global hotkeys.
			let handled = false;
			if (ev.keyCode === 89 /* y */) {
				// Make the URL absolute by adding the absolute 40-char commit ID
				// as the rev.
				if (this.state.commitID) {
					handled = true;
					(this.context as any).router.push(this._revSwitcherURL(this.state.commitID));
				}
			} else if (ev.keyCode === 85 /* u */) {
				// Remove the rev from the URL entirely.
				handled = true;
				(this.context as any).router.push(this._revSwitcherURL(null));
			} else if (ev.keyCode === 73 /* i */) {
				// Set the rev to be the repository's default branch.
				if (this.state.repoObj.DefaultBranch) {
					handled = true;
					(this.context as any).router.push(this._revSwitcherURL(this.state.repoObj.DefaultBranch));
				}
			}
			if (handled) {
				ev.preventDefault();
				ev.stopPropagation();
				return;
			}
		}

		if (!this.state.open) {
			return;
		}
		if (ev.keyCode === 27 /* ESC */) {
			this.setState(Object.assign({}, this.state, { open: false }));
		}
	}

	render(): JSX.Element | null {
		// Hide if cloning the repo, since we require the user to hard-reload. Seeing
		// the RevSwitcher would confuse them.
		if (this.state.isCloning) {
			return null;
		}

		let branches = this.state.branches.list(this.state.repo);
		if (this.state.branches.error(this.state.repo)) {
			branches = this._errorItem();
		} else if (!branches) {
			branches = this._loadingItem("branches");
		} else if (this.state.query) {
			branches = branches.filter((b) => fuzzysearch(this.state.query, b.Name));
		}
		if (branches.length === 0) {
			branches = this._emptyItem();
		}

		let tags = this.state.tags.list(this.state.repo);
		if (this.state.tags.error(this.state.repo)) {
			tags = this._errorItem();
		} else if (!tags) {
			tags = this._loadingItem("tags");
		} else if (this.state.query) {
			tags = tags.filter((t) => fuzzysearch(this.state.query, t.Name));
		}
		if (tags.length === 0) {
			tags = this._emptyItem();
		}

		let currentItem;
		if (branches instanceof Array) {
			branches.forEach((b) => {
				if (b.Name === this.state.effectiveRev) {
					currentItem = b;
				}
			});
		}
		if (tags instanceof Array) {
			tags.forEach((t) => {
				if (t.Name === this.state.effectiveRev) {
					currentItem = t;
				}
			});
		}

		if (branches instanceof Array) {
			branches = branches.map((b) => this._item(b.Name, b.Head));
		}
		if (tags instanceof Array) {
			tags = tags.map((t) => this._item(t.Name, t.CommitID));
		}

		let title;
		if (this.state.rev) {
			title = `Viewing revision: ${this._abbrevRev(this.state.rev)}`;
		}

		const sx = Object.assign({},
			{
				display: "inline-block",
				fontWeight: "normal",
				position: "relative",
			},
			typography.size[6],
		);

		return <div ref={(e) => this._wrapper = e} style={sx}>
			<span
				onClick={this._onToggleDropdown}
				style={{ cursor: "pointer" }}>
				<Base ml={1}>
					<DownMenu
						width={10}
						style={{ fill: colors.coolGray3() }}
					/>
				</Base>
			</span>
			<div style={{
				display: this.state.open ? "block" : "none",
				position: "absolute",
			}}>
				<Menu style={{minWidth: 320, paddingTop: whitespace[3]}}>
					<div>
						<Input block={true}
							domRef={(e) => this._input = e}
							type="text"
							style={{fontWeight: "normal"}}
							placeholder="Find branch or tag"
							onChange={this._onChangeQuery}/>
					</div>
					{this.state.rev && !currentItem && !this.state.query && this._item(this.state.rev, this.state.commitID)}
					<Heading level={7} color="gray" style={{marginTop: whitespace[3]}}>Branches</Heading>
					{branches}
					<Heading level={7} color="gray" style={{marginTop: whitespace[3]}}>Tags</Heading>
					{tags}
				</Menu>
			</div>
			<EventListener target={global.document} event="click" callback={this._onClickOutside} />
			<EventListener target={global.document} event="keydown" callback={this._onKeydown} />
		</div>;
	}
}
