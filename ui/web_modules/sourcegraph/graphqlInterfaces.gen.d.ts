// graphql typescript definitions

declare namespace GQL {
	interface IGraphQLResponseRoot {
		data?: IQuery;
		errors?: Array<IGraphQLResponseError>;
	}

	interface IGraphQLResponseError {
		message: string;            // Required for all errors
		locations?: Array<IGraphQLResponseErrorLocation>;
		[propName: string]: any;    // 7.2.2 says 'GraphQL servers may provide additional entries to error'
	}

	interface IGraphQLResponseErrorLocation {
		line: number;
		column: number;
	}

	/*
	  description: null
	*/
	interface ICommit {
		__typename: string;
		id: string;
		sha1: string;
		tree: ITree;
		file: IFile;
		languages: Array<string>;
	}

	/*
	  description: null
	*/
	interface ICommitState {
		__typename: string;
		commit: ICommit;
		cloneInProgress: boolean;
	}

	/*
	  description: null
	*/
	interface IContributor {
		__typename: string;
		login: string;
		avatarURL: string;
		contributions: number;
	}

	/*
	  description: null
	*/
	interface IDirectory {
		__typename: string;
		name: string;
		tree: ITree;
	}

	/*
	  description: null
	*/
	interface IFile {
		__typename: string;
		name: string;
		content: string;
		blame: Array<IHunk>;
	}

	/*
	  description: null
	*/
	interface IHunk {
		__typename: string;
		startLine: number;
		endLine: number;
		startByte: number;
		endByte: number;
		rev: string;
		name: string;
		email: string;
		date: string;
	}

	/*
	  description: null
	*/
	type Node = IRepository | ICommit;

	/*
	  description: null
	*/
	interface INode {
		__typename: string;
		id: string;
	}

	/*
	  description: null
	*/
	interface IQuery {
		__typename: string;
		root: IRoot;
		node: Node;
	}

	/*
	  description: null
	*/
	interface IRemoteRepository {
		__typename: string;
		uri: string;
		description: string;
		owner: string;
		name: string;
		httpCloneURL: string;
		language: string;
		fork: boolean;
		mirror: boolean;
		private: boolean;
		createdAt: string;
		pushedAt: string;
		vcsSyncedAt: string;
		contributors: Array<IContributor>;
	}

	/*
	  description: null
	*/
	interface IRepository {
		__typename: string;
		id: string;
		uri: string;
		description: string;
		commit: ICommitState;
		latest: ICommitState;
		defaultBranch: string;
		branches: Array<string>;
		tags: Array<string>;
		contributors: Array<IContributor>;
	}

	/*
	  description: null
	*/
	interface IRoot {
		__typename: string;
		repository: IRepository;
		remoteRepositories: Array<IRemoteRepository>;
		remoteStarredRepositories: Array<IRemoteRepository>;
	}

	/*
	  description: null
	*/
	interface ITree {
		__typename: string;
		directories: Array<IDirectory>;
		files: Array<IFile>;
	}
}
