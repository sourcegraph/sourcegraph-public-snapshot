// @flow

import type {LanguageID} from "sourcegraph/Language";

export type SearchSettings = {
	languages: Array<LanguageID>;
	scope: {
		popular: boolean;
		public: boolean;
		private: boolean;
	}
};
