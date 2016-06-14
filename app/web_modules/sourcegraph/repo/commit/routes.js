// @flow

import {rel} from "sourcegraph/app/routePatterns";
import urlTo from "sourcegraph/util/urlTo";
import {makeRepoRev} from "sourcegraph/repo";
import type {Route} from "react-router";

let _commitMain;

export const routes: Route = [
	{
		path: rel.commit,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				if (!_commitMain) {
					const withResolvedRepoRev = require("sourcegraph/repo/withResolvedRepoRev").default;
					_commitMain = withResolvedRepoRev(require("sourcegraph/repo/commit/CommitMain").default, true);
				}
				callback(null, {main: _commitMain});
			});
		},
	},
];

export function urlToRepoCommit(repo: string, rev: string): string {
	return urlTo("commit", {splat: makeRepoRev(repo, rev)});
}
