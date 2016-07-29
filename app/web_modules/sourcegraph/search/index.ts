import {LanguageID} from "sourcegraph/Language";

export type SearchScope = {
	popular: boolean;
	public: boolean;
	private: boolean;
	repo: boolean;
};

export const searchScopes = ["popular", "public", "private", "repo"];

export interface SearchSettings {
	languages: Array<LanguageID>;
	scope: SearchScope;
};
