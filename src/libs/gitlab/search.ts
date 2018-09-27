import { SearchFeature } from '../code_intelligence/search'

const getRepoInformation: SearchFeature['getRepoInformation'] = () => {
    const project = document.querySelector<HTMLElement>('.js-search-project-dropdown .dropdown-toggle-text')
    if (!project) {
        throw new Error('Unable to find project dropdown (search)')
    }

    const projectText = project.textContent || ''
    const parts = projectText
        .trim()
        .split(/\s/)
        .slice(1)

    const owner = parts[0]
    const repoName = parts[2]

    return {
        query: new URLSearchParams(window.location.search).get('search') || '',
        repoPath: `${window.location.host}/${owner}/${repoName}`,
    }
}

export const search: SearchFeature = {
    checkIsSearchPage: () =>
        window.location.pathname === '/search' &&
        new URLSearchParams(window.location.search).get('search_code') === 'true',
    getRepoInformation,
}
