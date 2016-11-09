import fuzzysearch from "fuzzysearch";
import * as debounce from "lodash/debounce";
import * as React from "react";
import * as Relay from "react-relay";
import {Link} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {Component, EventListener, isNonMonacoTextArea} from "sourcegraph/Component";
import {Heading, Input, Menu} from "sourcegraph/components";
import {Check, DownMenu} from "sourcegraph/components/symbols";
import {colors, typography, whitespace} from "sourcegraph/components/utils";
import "sourcegraph/repo/RepoBackend";
import {urlWithRev} from "sourcegraph/repo/routes";
import * as styles from "sourcegraph/repo/styles/RevSwitcher.css";

interface Props {
	repo: string;
	rev: string | null;
	commitID: string;
	repoObj?: any;
	isCloning: boolean;

	// to construct URLs
	routes: any[];
	routeParams: RouteParams;
}

interface State extends Props {
	open?: boolean;
	query?: any;
	effectiveRev?: string;
}

class RevSwitcherComponent extends Component<Props & {root: GQL.IRoot}, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	_input: any;
	_debouncedSetQuery: any;
	_wrapper: any;

	constructor(props: Props & {root: GQL.IRoot}) {
		super(props);
		this.state = {
			open: false,

			repo: props.repo,
			rev: props.rev,
			commitID: props.commitID,
			isCloning: props.isCloning,
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

	// abbrevRev shortens rev if it is an absolute commit ID.
	_abbrevRev(rev: string): string {
		return rev.length === 40 ? rev.substring(0, 12) : rev;
	}

	_item(name: string): JSX.Element {
		let isCurrent = name === this.state.effectiveRev;

		return (
			<div key={name} role="menu_item">
				<Link
					to={this._revSwitcherURL(name)}
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

	// _onKeydown controls the rev switcher and the page URL.
	_onKeydown(ev: KeyboardEvent): void {
		if (ev.defaultPrevented) {
			return;
		}

		// Don't trigger if there's a modifier key or if the cursor is focused
		// in an input field.
		const el = ev.target as HTMLElement;
		const tag = el.tagName;

		if (!(ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey) &&
			typeof document !== "undefined" && tag !== "INPUT" &&
			(tag !== "TEXTAREA" || !isNonMonacoTextArea(el)) &&
			tag !== "SELECT") {
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

		let branches = this.props.root.repository.branches;
		if (this.state.query) {
			branches = branches.filter((name) => fuzzysearch(this.state.query, name));
		}

		let tags = this.props.root.repository.tags;
		if (this.state.query) {
			tags = tags.filter((name) => fuzzysearch(this.state.query, name));
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
				<div style={{marginLeft: whitespace[1]}}>
					<DownMenu
						width={10}
						style={{ fill: colors.coolGray3() }}
					/>
				</div>
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
					<Heading level={7} color="gray" style={{marginTop: whitespace[3]}}>Branches</Heading>
					{branches.length === 0 && <li className={styles.disabled}>None found</li>}
					{branches.map((name) => this._item(name))}
					<Heading level={7} color="gray" style={{marginTop: whitespace[3]}}>Tags</Heading>
					{tags.length === 0 && <li className={styles.disabled}>None found</li>}
					{tags.map((name) => this._item(name))}
				</Menu>
			</div>
			<EventListener target={global.document.body} event="click" callback={this._onClickOutside} />
			<EventListener target={global.document.body} event="keydown" callback={this._onKeydown} />
		</div>;
	}
}

const RevSwitcherContainer = Relay.createContainer(RevSwitcherComponent, {
	initialVariables: {
		repo: "",
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repository(uri: $repo) {
					branches
					tags
				}
			}
		`,
	},
});

export const RevSwitcher = function(props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={RevSwitcherContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
	/>;
};
