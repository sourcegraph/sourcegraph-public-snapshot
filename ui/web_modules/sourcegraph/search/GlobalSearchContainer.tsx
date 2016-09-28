import {Location} from "history";
import * as React from "react";
import {EventListener} from "sourcegraph/Component";
import {Dismiss} from "sourcegraph/components/symbols/Dismiss";
import {colors} from "sourcegraph/components/utils";
import {GlobalSearchInput} from "sourcegraph/search/GlobalSearchInput";
import {SearchResultsPanel} from "sourcegraph/search/SearchResultsPanel";

interface Props {
	location: Location;
	style: Object;
}

interface State {
	open: boolean;
	focused: boolean;
	query: string;
}

export class GlobalSearchContainer extends React.Component<Props, State> {
	_container?: HTMLElement;
	_input?: HTMLInputElement;

	state: State = {
		open: false,
		focused: false,
		query: "",
	};

	constructor(props: Props) {
		super(props);

		this._handleGlobalHotkey = this._handleGlobalHotkey.bind(this);
		this._handleGlobalClick = this._handleGlobalClick.bind(this);
		this._handleReset = this._handleReset.bind(this);
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._handleChange = this._handleChange.bind(this);
		this._handleFocus = this._handleFocus.bind(this);
		this._handleBlur = this._handleBlur.bind(this);
	}

	_handleGlobalHotkey(ev: KeyboardEvent): void {
		if (ev.keyCode === 27 /* ESC */) {
			this.setState({open: false} as State);
		}
	}

	_handleGlobalClick(ev: Event): void {
		// Clicking outside of the open results panel should close it.
		if (this.state.open && (!this._container || !this._container.contains(ev.target as Node))) {
			this.setState({open: false} as State);
		}
	}

	_handleReset(ev: React.MouseEvent<HTMLButtonElement>): void {
		this.setState({focused: false, open: false, query: ""} as State);
		if (this._input) {
			this._input.value = "";
		}
	}

	_handleKeyDown(ev: React.KeyboardEvent<HTMLInputElement>): void {
		if (ev.keyCode === 27 /* ESC */) {
			this.setState({open: false} as State);
			if (this._input) {
				this._input.blur();
			}
		} else if (ev.keyCode === 13 /* Enter */) {
			// Close the search results menu AFTER the action has taken place on
			// the result (if a result was highlighted).
			setTimeout(() => this.setState({open: false} as State));
		}
	}

	_handleChange(ev: React.KeyboardEvent<HTMLInputElement>): void {
		const value = ev.currentTarget.value;
		if (value) {
			this.setState({query: value, open: true} as State);
		}
	}

	_handleFocus(ev: React.FocusEvent<HTMLInputElement>): void {
		const update = {focused: true, open: true};
		if (this._input && this._input.value) {
			(update as any).query = this._input.value;
		}
		this.setState(update as State);
	}

	_handleBlur(ev: React.FocusEvent<HTMLInputElement>): void {
		this.setState({focused: false} as State);
	}

	render(): JSX.Element | null {
		const containerSx = Object.assign({},
			{ position: "relative" },
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
				<GlobalSearchInput
					name="q"
					autoComplete="off"
					query={this.state.query || ""}
					domRef={e => this._input = e}
					onFocus={this._handleFocus}
					onBlur={this._handleBlur}
					onKeyDown={this._handleKeyDown}
					onClick={this._handleFocus}
					onChange={this._handleChange} />
					{this.state.open &&
						<button style={dismissBtnSx} type="reset" onClick={this._handleReset}>
							<Dismiss color={colors.coolGray3()}/>
						</button>
					}
				{this.state.open && <SearchResultsPanel query={this.state.query || ""} location={this.props.location} />}
				<EventListener target={global.document} event="keydown" callback={this._handleGlobalHotkey} />
				<EventListener target={global.document} event="click" callback={this._handleGlobalClick} />
			</div>
		);
	}
}
