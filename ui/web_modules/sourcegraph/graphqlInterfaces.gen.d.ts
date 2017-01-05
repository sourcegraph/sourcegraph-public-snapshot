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
		tree: ITree | null;
		file: IFile | null;
		languages: Array<string>;
	}

	/*
	  description: null
	*/
	interface ICommitState {
		__typename: string;
		commit: ICommit | null;
		cloneInProgress: boolean;
	}

	/*
	  description: null
	*/
	interface IDefinition {
		__typename: string;
		globalReferences: Array<IRefFields>;
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
		definition: IDefinition | null;
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
		message: string;
		gravatarHash: string;
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
		node: Node | null;
	}

	/*
	  description: null
	*/
	interface IRefFields {
		__typename: string;
		refLocation: IRefLocation | null;
		uri: IURI | null;
	}

	/*
	  description: null
	*/
	interface IRefLocation {
		__typename: string;
		startLineNumber: number;
		startColumn: number;
		endLineNumber: number;
		endColumn: number;
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
		language: string;
		fork: boolean;
		private: boolean;
		createdAt: string;
		pushedAt: string;
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
	}

	/*
	  description: null
	*/
	interface IRoot {
		__typename: string;
		repository: IRepository | null;
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

	/*
	  description: null
	*/
	interface IURI {
		__typename: string;
		host: string;
		fragment: string;
		path: string;
		query: string;
		scheme: string;
	}
}
