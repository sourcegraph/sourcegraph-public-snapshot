import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import RepoStore from "sourcegraph/repo/RepoStore";

// preloadStores resets all stores and preloads them with the provided data.
export default function preloadStores(data) {
	if (!data || typeof data !== "object") throw new Error("data must be an object");
	RepoStore.reset(data.RepoStore || {});
	BlobStore.reset(data.BlobStore || {});
	DefStore.reset(data.DefStore || {});
}
