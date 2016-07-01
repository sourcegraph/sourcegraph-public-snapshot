// @flow

import type {LanguageID} from "sourcegraph/Language";

export type SearchScope = {
	popular: boolean;
	public: boolean;
	private: boolean;
	repo: boolean;
};

export type SearchSettings = {
	languages: Array<LanguageID>;
	scope: SearchScope,
};
