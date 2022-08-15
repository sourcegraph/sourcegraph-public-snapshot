export enum SectionID {
    SEARCH_REFERENCE = 'reference',
    SEARCH_TYPES = 'types',
    DYNAMIC_FILTERS = 'filters', // Deprecated
    LANGUAGES = 'languages',
    REPOSITORIES = 'repositories',
    FILE_TYPES = 'file-types',
    OTHER = 'other',
    SEARCH_SNIPPETS = 'snippets',
    QUICK_LINKS = 'quicklinks',
    REVISIONS = 'revisions',
}

export enum NoResultsSectionID {
    SEARCH_BAR = 'search-bar',
    LITERAL_SEARCH = 'literal-search',
    COMMON_PROBLEMS = 'common-problems',
    VIDEOS = 'videos',
}

export type SidebarTabID = 'filters'
