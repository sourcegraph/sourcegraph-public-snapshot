import {urlToBlob} from "sourcegraph/blob/routes";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {defaultFetch} from "sourcegraph/util/xhr";

import {Category, Result} from "sourcegraph/search/modal/SearchContainer";

import {TreeStore} from "sourcegraph/tree/TreeStore";

const fetch = singleflightFetch(defaultFetch);

// Update category finds the results for a category and calls the given callback
// with the results.
export function updateCategory(category: Category, repo: string, commitID: string, query: string, callback: (results: Result[]) => void): void {
	if (category === Category.repository) {
		updateRepos(query, callback);
		return;
	} else if (category === Category.file) {
		callback(fileSearch(repo, commitID, query));
		return;
	}
}

function updateRepos(query: string, callback: (results: Result[]) => void): void {
	const url = `/.api/search-repos?Query=${query}`;
	fetch(url)
	.then(resp => resp.json())
	.then(data => data.Repos === undefined ? [] :
		data.Repos.map(({URI}) => ({title: URI, URLPath: `/${URI}`})))
	.then(callback)
	.catch(error => console.error(error));
}

function fileSearch(repo: string, commit: string, query: string): Result[] {
	let lowerQuery = query.toLowerCase();

	let results = new Array();
	if (lowerQuery === "") {
		return results;
	}
	let fileList = TreeStore.fileLists.get(repo, commit);
	if (fileList === null) {
		return results;
	}
	fileList.Files.forEach((file, n) => {
		let index = file.toLowerCase().indexOf(lowerQuery);
		if (index === -1) {
			return;
		}
		results.push({
			title: file,
			index: index,
			length: lowerQuery.length,
			URLPath: urlToBlob(repo, null, file),
		});
	});
	return results;
}
