// @flow

import type {RepoRevSpec} from "sourcegraph/repo";

export type DeltaSpec = {
	Base: RepoRevSpec;
	Head: RepoRevSpec;
};

export type DeltaFiles = {
	Ds: DeltaSpec;
	Opt?: ?Object;
};
