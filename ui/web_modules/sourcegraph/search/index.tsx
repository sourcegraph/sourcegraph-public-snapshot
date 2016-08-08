export interface SearchScope {
	popular: boolean;
	public: boolean;
	private: boolean;
	repo: boolean;
};

export const searchScopes = ["popular", "public", "private", "repo"];

export interface SearchSettings {
	languages: string[];
	scope: SearchScope;
};
