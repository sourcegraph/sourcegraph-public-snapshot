import * as csstips from "csstips";
import * as React from "react";
import * as ChevronDown from "react-icons/lib/fa/chevron-down";
import * as CloseIcon from "react-icons/lib/fa/close";
import * as Rx from "rxjs";
import { expandActiveInactive, getSearchParamsFromLocalStorage, getSearchParamsFromURL, parseRepoList } from "sourcegraph/search";
import { setState as setSearchState, store as searchStore } from "sourcegraph/search/store";
import { getCurrent as getActiveRepos } from "sourcegraph/util/activeRepos";
import { inputBackgroundColor, white } from "sourcegraph/util/colors";
import { isBlob, isSearchResultsPage, parseBlob } from "sourcegraph/util/url";
import { style } from "typestyle";

namespace Styles {
	export const container = style(csstips.horizontal, csstips.center, { backgroundColor: inputBackgroundColor, color: white, padding: "8px 12px", marginLeft: "8px", cursor: "pointer", height: "32px", fontSize: "13px" });
	export const chevron = style({ fontSize: "10px", marginLeft: "16px" });
}

interface State {
	scope: string;
	showAdvancedSearch?: boolean;
}

export class AdvancedSearchToggle extends React.Component<{}, State> {
	subscription: Rx.Subscription;

	constructor(props: {}) {
		super(props);
		const url = parseBlob();
		const loc = window.location;
		const searchParams = isSearchResultsPage(loc) ? getSearchParamsFromURL(loc.href) : getSearchParamsFromLocalStorage();
		let repoList = parseRepoList(searchParams.repos);
		const activeRepos = getActiveRepos(); // TODO(john): update when new results are fetched
		if (activeRepos) {
			repoList = expandActiveInactive(repoList, activeRepos);
		}

		this.state = {
			scope: isBlob(url) ? url.uri!.substr("github.com/".length) /* TODO(john): fix <-- that */ : `${repoList.length} repositor${repoList.length === 1 ? "y" : "ies"}`,
			showAdvancedSearch: searchStore.getValue().showAdvancedSearch,
		};
	}

	componentDidMount(): void {
		this.subscription = searchStore.subscribe((state) => {
			this.setState({ ...state, scope: this.state.scope } as any);
		});
	}

	componentWillUnmount(): void {
		if (this.subscription) {
			this.subscription.unsubscribe();
		}
	}

	onClick(): void {
		setSearchState({ ...searchStore.getValue(), showAdvancedSearch: !searchStore.getValue().showAdvancedSearch });
	}

	render(): JSX.Element | null {
		return <div className={Styles.container} onClick={() => this.onClick()}>
			<span>Current scope: {this.state.scope}</span>
			{this.state.showAdvancedSearch ? <CloseIcon className={Styles.chevron} /> : <ChevronDown className={Styles.chevron} />}
		</div>;
	}
}
