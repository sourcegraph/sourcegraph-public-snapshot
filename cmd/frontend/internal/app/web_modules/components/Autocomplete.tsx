import * as scrollIntoView from "dom-scroll-into-view";
import * as _ from "lodash";
import * as React from "react";
import * as ReactDOM from "react-dom";

interface Props {
	ItemView: any;
	onEscape: () => void;
	onChange: (query: string) => void;
	onSelect: (item: any) => void;
	emptyMessage: string;
	inputProps?: any;
	className: string;
	inputClassName: string;
	autocompleteResultsClassName: string;
	emptyClassName: string;
	onMount?: () => void;
}

interface State {
	isOpen: boolean;
	loading: boolean;
	items: any[];
	index: number;
}

export class Autocomplete extends React.Component<Props, State> {
	scrollingIntoView: boolean;
	keyDownHandlers: any;
	onChangeInput: any;

	constructor(props: Props) {
		super(props);
		this.state = {
			// true to show the options, false otherwise
			isOpen: false,

			// true to show the loading indicator
			loading: false,

			// the options to display
			items: [],

			// the index of the highlighted option
			index: -1,
		};
		// true if the user is scrolling with keys, to ignore mouse events until done
		this.scrollingIntoView = false;
		this.keyDownHandlers = {
			ArrowDown() {
				if (!this.state.isOpen && this.state.items.length) {
					this.setState({
						isOpen: true,
					});
				} else {
					this.moveSelectedOption(1);
				}
			},

			ArrowUp() {
				this.moveSelectedOption(-1);
			},

			Enter() {
				const { isOpen, index } = this.state;
				if (isOpen && index > -1) {
					this.onSelectIndex(index);
				}
			},

			Escape() {
				// this.hideItems();
				this.props.onEscape();
			},

			PageUp() {
				this.state.isOpen && this.moveSelectedOption(-10);
			},

			PageDown() {
				this.state.isOpen && this.moveSelectedOption(10);
			},

			End() {
				this.state.isOpen && this.setState({
					index: (this.state.items.length || 0) - 1,
					isOpen: true,
				});
			},

			Home() {
				this.state.isOpen && this.setState({
					index: 0,
					isOpen: true,
				});
			},
		};

		_.bindAll(this, ["hideItems", "onSelectIndex", "onKeyDown", "onMouseOver", "onClickItem"]);
		this.onChangeInput = _.throttle(function() {
			this.setState({
				loading: true,
			});
			this.props.onChange((ReactDOM.findDOMNode(this.refs.input) as any).value);
		}.bind(this), 200, { leading: false });

	}

	setItems(items: any[]): void {
		this.setState({
			isOpen: true,
			loading: false,
			items: items,
			index: items.length ? 0 : -1,
		});
	}

	hideItems(): void {
		this.setState({
			isOpen: false,
			loading: false,
		});
	}

	componentWillUpdate(_, { isOpen }): void {
		const prevIsOpen = this.state.isOpen;
		if (prevIsOpen && !isOpen) {
			// document.removeEventListener('click', this.hideItems);
		} else if (!prevIsOpen && isOpen) {
			// document.addEventListener("click", this.hideItems);
		}
	}

	componentDidUpdate(): void {
		if (this.state.isOpen === true) {
			const selected = document.getElementsByClassName("autocomplete-li selected");
			if (selected.length) {
				this.scrollingIntoView = true;
				scrollIntoView(selected[0],
					selected[0].parentElement,
					{ onlyScrollIfNeeded: true },
				);
			}
		}
	}

	componentDidMount(): void {
		if (this.props.onMount) {
			this.props.onMount();
		}
	}

	onMouseOver(e: any): void {
		if (this.scrollingIntoView) {
			this.scrollingIntoView = false;
			return;
		}
		let element = e.target;
		let _index;
		do {
			_index = element.getAttribute("data-index");
			element = element.parentElement;
		} while (!_index && element);
		_index && this.setState({
			index: +_index,
		});
	}

	onClickItem(e) {
		const _index = +e.currentTarget.getAttribute("data-index");
		this.onSelectIndex(_index);
	}

	onSelectIndex(index: number): void {
		const item = this.state.items[index];
		const newInputValue = this.props.onSelect(item);
		const $input = this.refs.input as any;
		// this.hideItems();
		$input.value = newInputValue || "";
		newInputValue && $input.select();
	}

	onKeyDown(event) {
		const handler = this.keyDownHandlers[event.key];
		if (handler) {
			event.preventDefault();
			handler.call(this, event);
		}
	}

	// select the next or previous option
	// @param delta +1 or -1 to move to the next or previous choice
	moveSelectedOption(delta) {
		let { index, items } = this.state;
		if (!items.length) {
			index = -1;
		} else {
			let index = ((index || 0) + delta) % items.length;
			if (index < 0) {
				index = 0;
			}
		}
		this.setState({
			index: index,
			isOpen: true,
		});

	}

	renderItems() {
		const { items, index, isOpen } = this.state;
		const $empty = items && items.length ? undefined :
			<div className={this.props.emptyClassName}>{this.props.emptyMessage}</div>;
		return !isOpen ? undefined : (
			<div className={this.props.autocompleteResultsClassName} onMouseOver={this.onMouseOver}>
				{$empty || items.map((item, _index) => {
					return (
						<div
							className={"autocomplete-li" + (index == _index ? " selected" : "")}
							key={_index}
							onClick={this.onClickItem}
							data-index={_index}
						>
							<this.props.ItemView item={item} highlighted={index == _index} />
						</div>
					);
				})}
			</div>
		);
	}

	render() {
		return (
			<div className={this.props.className} id="autocomplete">
				<input
					autoFocus
					type="text"
					ref="input"
					className={this.props.inputClassName}
					autoComplete="off"
					aria-autocomplete="list"
					{... this.props.inputProps}
					onChange={this.onChangeInput}
					onKeyDown={this.onKeyDown}
				/>
				{this.renderItems()}
				{this.state.loading ? <div className="loading"></div> : undefined}
			</div>
		);
	}
}
