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
  interface IBlob {
    __typename: string;
    bytes: string;
  }

  /*
    description: null
  */
  interface ICommit {
    __typename: string;
    id: string;
    sha1: string;
    tree: ITree;
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
    content: IBlob;
  }

  /*
    description: null
  */
  type Node = IRepository | ICommit;

  /*
    description: null
  */
  interface INode extends IRepository, ICommit {
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
  interface IRepository {
    __typename: string;
    id: string;
    uri: string;
    commit: ICommit;
    latest: ICommit;
  }

  /*
    description: null
  */
  interface IRoot {
    __typename: string;
    repository: IRepository;
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
