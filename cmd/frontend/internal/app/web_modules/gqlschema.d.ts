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
  interface IQuery {
    __typename: string;
    root: IRoot;
    node: INode | null;
  }

  /*
    description: null
  */
  interface IRoot {
    __typename: string;
    organization: IOrganization | null;
    repository: IRepository | null;
    repositories: Array<IRepository>;
    remoteRepositories: Array<IRemoteRepository>;
    remoteStarredRepositories: Array<IRemoteRepository>;
    symbols: Array<ISymbol>;
    currentUser: IUser | null;
    activeRepos: IActiveRepoResults;
    searchRepos: ISearchResults;
    searchProfiles: Array<ISearchProfile>;
    revealCustomerCompany: ICompanyProfile | null;
    threads: Array<IThread>;
  }

  /*
    description: null
  */
  interface IOrganization {
    __typename: string;
    login: string;
    githubId: number;
    email: string;
    name: string;
    avatarURL: string;
    description: string;
    collaborators: number;
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
    starsCount: number | null;
    forksCount: number | null;
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
    listTotalRefs: ITotalRefList;
    gitCmdRaw: string;
  }

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
  interface ICommitState {
    __typename: string;
    commit: ICommit | null;
    cloneInProgress: boolean;
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
  interface ITree {
    __typename: string;
    directories: Array<IDirectory>;
    files: Array<IFile>;
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
    blameRaw: string;
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
  interface ISignature {
    __typename: string;
    person: IPerson | null;
    date: string;
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
  interface IDependencyReferences {
    __typename: string;
    dependencyReferenceData: IDependencyReferencesData;
    repoData: IRepoDataMap;
  }

  /*
    description: null
  */
  interface IDependencyReferencesData {
    __typename: string;
    references: Array<IDependencyReference>;
    location: IDepLocation;
  }

  /*
    description: null
  */
  interface IDependencyReference {
    __typename: string;
    dependencyData: string;
    repoId: number;
    hints: string;
  }

  /*
    description: null
  */
  interface IDepLocation {
    __typename: string;
    location: string;
    symbol: string;
  }

  /*
    description: null
  */
  interface IRepoDataMap {
    __typename: string;
    repos: Array<IRepository>;
    repoIds: Array<number>;
  }

  /*
    description: null
  */
  interface IRevState {
    __typename: string;
    commit: ICommit | null;
    cloneInProgress: boolean;
  }

  /*
    description: null
  */
  interface ITotalRefList {
    __typename: string;
    repositories: Array<IRepository>;
    total: number;
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
  interface IUser {
    __typename: string;
    githubInstallations: Array<IInstallation>;
  }

  /*
    description: null
  */
  interface IInstallation {
    __typename: string;
    login: string;
    githubId: number;
    installId: number;
    type: string;
    avatarURL: string;
  }

  /*
    description: null
  */
  interface IActiveRepoResults {
    __typename: string;
    active: Array<string>;
    inactive: Array<string>;
  }

  /*
    description: null
  */
  interface ISearchQuery {
    __typename: string;
    pattern: string;
    isRegExp: boolean;
    isWordMatch: boolean;
    isCaseSensitive: boolean;
    fileMatchLimit: number;
    includePattern: string | null;
    excludePattern: string | null;
  }

  /*
    description: null
  */
  interface IRepositoryRevision {
    __typename: string;
    repo: string;
    rev: string | null;
  }

  /*
    description: null
  */
  interface ISearchResults {
    __typename: string;
    results: Array<IFileMatch>;
    limitHit: boolean;
    cloning: Array<string>;
    missing: Array<string>;
  }

  /*
    description: null
  */
  interface IFileMatch {
    __typename: string;
    resource: string;
    lineMatches: Array<ILineMatch>;
    limitHit: boolean;
  }

  /*
    description: null
  */
  interface ILineMatch {
    __typename: string;
    preview: string;
    lineNumber: number;
    offsetAndLengths: Array<Array<number>>;
    limitHit: boolean;
  }

  /*
    description: null
  */
  interface ISearchProfile {
    __typename: string;
    name: string;
    description: string | null;
    repositories: Array<IRepository>;
  }

  /*
    description: null
  */
  interface ICompanyProfile {
    __typename: string;
    ip: string;
    domain: string;
    fuzzy: boolean;
    company: ICompanyInfo;
  }

  /*
    description: null
  */
  interface ICompanyInfo {
    __typename: string;
    id: string;
    name: string;
    legalName: string;
    domain: string;
    domainAliases: Array<string>;
    url: string;
    site: ISiteDetails;
    category: ICompanyCategory;
    tags: Array<string>;
    description: string;
    foundedYear: string;
    location: string;
    logo: string;
    tech: Array<string>;
  }

  /*
    description: null
  */
  interface ISiteDetails {
    __typename: string;
    url: string;
    title: string;
    phoneNumbers: Array<string>;
    emailAddresses: Array<string>;
  }

  /*
    description: null
  */
  interface ICompanyCategory {
    __typename: string;
    sector: string;
    industryGroup: string;
    industry: string;
    subIndustry: string;
  }

  /*
    description: null
  */
  interface IThread {
    __typename: string;
    id: number;
    file: string;
    revision: string;
    startLine: number;
    endLine: number;
    startCharacter: number;
    endCharacter: number;
    createdAt: string;
    comments: Array<IComment>;
  }

  /*
    description: null
  */
  interface IComment {
    __typename: string;
    id: number;
    contents: string;
    createdAt: string;
    updatedAt: string;
    authorName: string;
    authorEmail: string;
  }

  /*
    description: null
  */
  interface IMutation {
    __typename: string;
    createThread: IThread;
    addCommentToThread: IThread;
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
  interface IURI {
    __typename: string;
    host: string;
    fragment: string;
    path: string;
    query: string;
    scheme: string;
  }
}

// tslint:enable
