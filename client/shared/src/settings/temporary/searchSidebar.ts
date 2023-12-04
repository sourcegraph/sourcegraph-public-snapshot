export enum SectionID {
    GROUPED_BY = 'grouped-by',
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
    SEVERITY = 'severity',
}

export enum NoResultsSectionID {
    SEARCH_BAR = 'search-bar',
}

export type SidebarTabID = 'filters'
