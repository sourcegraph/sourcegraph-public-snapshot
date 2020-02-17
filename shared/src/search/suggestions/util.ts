export enum SuggestionTypes {
    filters = 'filters',
    repo = 'repo',
    repogroup = 'repogroup',
    repohasfile = 'repohasfile',
    repohascommitafter = 'repohascommitafter',
    file = 'file',
    type = 'type',
    case = 'case',
    lang = 'lang',
    fork = 'fork',
    archived = 'archived',
    count = 'count',
    timeout = 'timeout',
    dir = 'dir',
    symbol = 'symbol',
    before = 'before',
    after = 'after',
    author = 'author',
    message = 'message',
    content = 'content',
}

export const suggestionTypeKeys: SuggestionTypes[] = Object.keys(SuggestionTypes) as SuggestionTypes[]
