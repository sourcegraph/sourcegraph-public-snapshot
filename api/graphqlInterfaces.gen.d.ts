// tslint:disable
// graphql typescript definitions

declare namespace GQL {
	interface IGraphQLResponseRoot {
		data?: IQuery | IMutation;
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
		textSearch: Array<IFileMatch>;
		file: IFile | null;
		languages: Array<string>;
	}

	/*
	  description: null
	*/
	interface ICommitInfo {
		__typename: string;
		rev: string;
		author: ISignature | null;
		committer: ISignature | null;
		message: string;
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
	interface IDependencyReferences {
		__typename: string;
		data: string;
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
		commits: Array<ICommitInfo>;
		dependencyReferences: IDependencyReferences;
	}

	/*
	  description: null
	*/
	interface IFileMatch {
		__typename: string;
		path: string;
		lineMatches: Array<ILineMatch>;
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
		author: ISignature | null;
		message: string;
	}

	/*
	  description: null
	*/
	interface ILineMatch {
		__typename: string;
		preview: string;
		lineNumber: number;
	}

	/*
	  description: null
	*/
	interface IMutation {
		__typename: string;
		cancelSubscription: boolean;
		updatePaymentSource: boolean;
		subscribeOrg: boolean;
		startOrgTrial: boolean;
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
	interface IOrganization {
		__typename: string;
		name: string;
		avatarURL: string;
		description: string;
		collaborators: number;
	}

	/*
	  description: null
	*/
	interface IPerson {
		__typename: string;
		name: string;
		email: string;
		gravatarHash: string;
	}

	/*
	  description: null
	*/
	interface IPlan {
		__typename: string;
		name: string;
		cost: number;
		seats: number | null;
		renewalDate: number | null;
		organization: IOrganization | null;
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
		language: string;
		fork: boolean;
		private: boolean;
		createdAt: string;
		pushedAt: string;
		commit: ICommitState;
		revState: IRevState;
		latest: ICommitState;
		lastIndexedRevOrLatest: ICommitState;
		defaultBranch: string;
		branches: Array<string>;
		tags: Array<string>;
		expirationDate: number | null;
	}

	/*
	  description: null
	*/
	interface IRevState {
		__typename: string;
		zapRev: IZapRev | null;
		commit: ICommit | null;
		cloneInProgress: boolean;
	}

	/*
	  description: null
	*/
	interface IRoot {
		__typename: string;
		repository: IRepository | null;
		repositories: Array<IRepository>;
		remoteRepositories: Array<IRemoteRepository>;
		remoteStarredRepositories: Array<IRemoteRepository>;
		symbols: Array<ISymbol>;
		currentUser: IUser | null;
	}

	/*
	  description: null
	*/
	interface ISignature {
		__typename: string;
		person: IPerson | null;
		date: string;
	}

	/*
	  description: null
	*/
	interface ISymbol {
		__typename: string;
		repository: IRepository;
		path: string;
		line: number;
		character: number;
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

	/*
	  description: null
	*/
	interface IUser {
		__typename: string;
		githubOrgs: Array<IOrganization>;
		paymentPlan: IPlan;
	}

	/*
	  description: null
	*/
	interface IZapRev {
		__typename: string;
		ref: string;
		base: string;
		branch: string;
	}
}

// tslint:enable
