import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import RepoStore from "sourcegraph/repo/RepoStore";
import TreeStore from "sourcegraph/tree/TreeStore";
import SearchStore from "sourcegraph/search/SearchStore";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import BuildStore from "sourcegraph/build/BuildStore";
import UserStore from "sourcegraph/user/UserStore";

export default function dumpStores() {
	return {
		RepoStore,
		BlobStore,
		DefStore,
		TreeStore,
		SearchStore,
		DashboardStore,
		BuildStore,
		UserStore,
	};
}
