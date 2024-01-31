export enum FilterKind {
    SymbolKind = 'symbol type',
    Language = 'lang',
    Author = 'author',
    Repository = 'repo',
    CommitDate = 'commit date',
    File = 'file',
    Utility = 'utility',

    // Synthetic filter, lives only on the client
    Count = 'count',
}
