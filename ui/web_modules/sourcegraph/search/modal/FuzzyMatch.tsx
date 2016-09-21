import "string_score";

import {Result} from "sourcegraph/search/modal/SearchContainer";

interface Scorable {
	score?: number;
	title: any;
}

// Sort a category according to the file.
export function fuzzyRankResults(query: string, results: Result[]): Result[] {
	results.forEach((r: Result & Scorable) => r.score = r.title.score(query, 0.5));
	if (query !== "") {
		results = results.filter((r: Scorable) => r.score > .25);
	}
	return results.sort((a: Scorable, b: Scorable) => b.score - a.score);
}
