import * as React from "react";

interface Props {
	className?: string;
	style?: React.CSSProperties;
	children?: React.ReactNode[];
	offset?: number;
}

export class Affix extends React.Component<Props, {}> {
	_affix: {
		offsetTop: number,
		style: any,
		clientWidth: number,
	};

	componentDidMount(): void {
		const initialOffset = this._getOffset();
		window.addEventListener("scroll", () => this._affixEl(initialOffset));
	}

	componentWillUnmount(): void {
		const initialOffset = this._getOffset();
		window.removeEventListener("scroll", () => this._affixEl(initialOffset));
	}

	_getOffset(): number {
		return this._affix.offsetTop;
	}

	_affixEl(initialOffset: number): any {
		if (!this._affix) {
			return;
		}

		const currentWidth = this._affix.clientWidth;
		if (initialOffset <= window.scrollY) {
			this._affix.style.width = `${currentWidth}px`;
			this._affix.style.position = "fixed";
			this._affix.style.top = `${this.props.offset}px`;
		} else if (initialOffset > window.scrollY) {
			this._affix.style.position = "relative";
		}
	}

	render(): JSX.Element {
		const {className, style, children} = this.props;
		return <div className={className} style={style}>
			<div ref={(el) => this._affix = el }>{children}</div>
		</div>;
	}
}
