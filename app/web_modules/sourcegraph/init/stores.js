// @flow

import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import RepoStore from "sourcegraph/repo/RepoStore";
import TreeStore from "sourcegraph/tree/TreeStore";
import SearchStore from "sourcegraph/search/SearchStore";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import BuildStore from "sourcegraph/build/BuildStore";
import CoverageStore from "sourcegraph/admin/CoverageStore";
import UserStore from "sourcegraph/user/UserStore";

// allStores is a function because there is a cyclic dependency between
// UserStore and this module, so UserStore is null at eval-time.
const allStores = () => ({
	BlobStore,
	DefStore,
	RepoStore,
	TreeStore,
	SearchStore,
	DashboardStore,
	BuildStore,
	CoverageStore,
	UserStore,
});

// forEach calls f(store, name) for all stores.
export function forEach(f: (store: Object, name: string) => void): void {
	const stores = allStores();
	Object.keys(stores).forEach((key) => {
		f(stores[key], key);
	});
}

// reset resets all stores with the provided data. If null is provided,
// then the stories are cleared.
export function reset(data: ?Object): void {
	forEach((store, name) => {
		store.reset(data ? data[name] : null);
	});
}
