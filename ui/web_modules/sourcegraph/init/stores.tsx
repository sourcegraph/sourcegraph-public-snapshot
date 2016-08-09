// tslint:disable

import {BlobStore} from "sourcegraph/blob/BlobStore";
import {DefStore} from "sourcegraph/def/DefStore";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import {SearchStore} from "sourcegraph/search/SearchStore";
import {BuildStore} from "sourcegraph/build/BuildStore";
import {UserStore} from "sourcegraph/user/UserStore";

// allStores is a function because there is a cyclic dependency between
// UserStore and this module, so UserStore is null at eval-time.
const allStores = () => ({
	BlobStore,
	DefStore,
	RepoStore,
	TreeStore,
	SearchStore,
	BuildStore,
	UserStore,
});

// forEach calls f(store, name) for all stores.
export function forEach(f: (store: any, name: string) => void): void {
	const stores = allStores();
	Object.keys(stores).forEach((key) => {
		f(stores[key], key);
	});
}

// reset resets all stores.
export function reset(): void {
	forEach((store, name) => {
		store.reset();
	});
}
