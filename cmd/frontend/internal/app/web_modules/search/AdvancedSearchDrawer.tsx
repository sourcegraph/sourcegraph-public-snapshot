import { getSearchPath, SearchParams } from "app/search";
import { setState as setSearchState, store as searchStore } from "app/search/store";
import { inputBackgroundColor, normalFontColor, primaryBlue, referencesBackgroundColor, white } from "app/util/colors";
import { isBlobPage, parseBlob } from "app/util/url";
import * as csstips from "csstips";
import * as React from "react";
import * as CheckboxFilled from "react-icons/lib/md/check-box";
import * as CheckboxOutline from "react-icons/lib/md/check-box-outline-blank";
import * as SearchIcon from "react-icons/lib/md/search";
import * as Rx from "rxjs";
import { style } from "typestyle";

namespace Styles {
	const padding = "10px";
	const borderRadius = "2px";
	const rowHeight = "32px";
	const input = { backgroundColor: inputBackgroundColor, padding, border: "none", color: white, fontSize: "14px", $nest: { "&::placeholder": { color: normalFontColor }, "&:focus": { outline: "none" } } };

	export const icon = style({ fontSize: "18px", marginRight: "8px" });

	export const container = style(csstips.horizontal, csstips.center, { backgroundColor: referencesBackgroundColor, color: normalFontColor, padding: "8px 12px", fontSize: "13px" });

	export const repoArea = style(csstips.flex3, { maxWidth: "50%", height: "64px" });
	export const reposInput = style(input, { borderRadius, minHeight: "100%", width: "100%", maxHeight: "100%", minWidth: "100%", maxWidth: "100%" });

	export const filesSection = style(csstips.flex2, { marginLeft: "16px" });
	export const filesInput = style(input, { marginTop: "8px", borderRadius, height: rowHeight, width: "100%" });

	export const filtersSection = style(csstips.flex1, csstips.vertical, { marginLeft: "16px" });
	export const filter = style(csstips.flex, csstips.horizontal, csstips.center, { cursor: "pointer", marginRight: "16px", userSelect: "none" });

	export const searchButton = style(csstips.horizontal, csstips.center, csstips.content, { backgroundColor: primaryBlue, height: rowHeight, padding, borderRadius, color: `${white} !important`, textDecoration: "none" });

}

export class AdvancedSearchDrawer extends React.Component<{}, SearchParams> {
	subscription: Rx.Subscription;

	constructor(props: {}) {
		super(props);
		this.state = searchStore.getValue();
	}

	componentDidMount(): void {
		this.subscription = searchStore.subscribe((state) => {
			this.setState(state);
		});

		if (isBlobPage()) {
			setSearchState({ ...searchStore.getValue(), repos: parseBlob().uri! });
		}
	}

	componentWillUnmount(): void {
		if (this.subscription) {
			this.subscription.unsubscribe();
		}
	}

	render(): JSX.Element | null {
		return <div className={Styles.container}>
			<div className={Styles.repoArea}>
				<textarea className={Styles.reposInput} value={this.state.repos} onChange={(e) => {
					setSearchState({ ...searchStore.getValue(), repos: e.target.value });
				}} />
			</div>
			<div className={Styles.filesSection}>
				<div>Files to include</div>
				<input className={Styles.filesInput} value={this.state.files} placeholder="example: *.go" onChange={(e) => {
					setSearchState({ ...searchStore.getValue(), files: e.target.value });
				}} />
			</div>
			<div className={Styles.filtersSection}>
				{
					[{ key: "matchCase", label: "Match case" }, { key: "matchWord", label: "Match word" }, { key: "matchRegex", label: "Regex" }]
						.map((filter, i) => {
							const clickHandler = () => {
								setSearchState({ ...searchStore.getValue(), [filter.key]: !this.state[filter.key] });
							};
							return <div key={i} className={Styles.filter} onClick={clickHandler}>
								{this.state[filter.key] ? <CheckboxFilled className={Styles.icon} /> : <CheckboxOutline className={Styles.icon} />}
								{filter.label}
							</div>;
						})
				}
			</div>
			<a className={Styles.searchButton} href={getSearchPath(this.state)}>
				<SearchIcon className={Styles.icon} />
				Search code
				</a>
		</div>;
	}
}
