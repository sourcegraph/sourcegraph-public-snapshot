import { FileInfo } from '../shared/codeHost'
import { getCommitIDFromPermalink } from './scrape'
import { getDiffFileName, getDiffResolvedRev, getFilePath, parseURL } from './util'

export const resolveDiffFileInfo = (codeView: HTMLElement): FileInfo => {
    const { rawRepoName } = parseURL()
    const { headFilePath, baseFilePath } = getDiffFileName(codeView)
    if (!headFilePath) {
        throw new Error('cannot determine file path')
    }
    const diffResolvedRev = getDiffResolvedRev(codeView)
    if (!diffResolvedRev) {
        throw new Error('cannot determine delta info')
    }
    const { headCommitID, baseCommitID } = diffResolvedRev
    return {
        rawRepoName,
        filePath: headFilePath,
        commitID: headCommitID,
        rev: headCommitID,
        baseRawRepoName: rawRepoName,
        baseFilePath,
        baseCommitID,
        baseRev: baseCommitID,
    }
}

export const resolveFileInfo = (): FileInfo => {
    const parsedURL = parseURL()
    if (parsedURL.pageType !== 'blob' && parsedURL.pageType !== 'tree') {
        throw new Error(`Current URL does not match a blob or tree url: ${window.location.href}`)
    }
    const { revAndFilePath, rawRepoName } = parsedURL

    const filePath = getFilePath()
    const filePathWithLeadingSlash = filePath.startsWith('/') ? filePath : `/${filePath}`
    if (!revAndFilePath.endsWith(filePathWithLeadingSlash)) {
        throw new Error(
            `The file path ${filePathWithLeadingSlash} should always be a suffix of revAndFilePath ${revAndFilePath}, but isn't in this case.`
        )
    }
    return {
        rawRepoName,
        filePath,
        commitID: getCommitIDFromPermalink(),
        rev: revAndFilePath.slice(0, -filePathWithLeadingSlash.length),
    }
}

const COMMIT_HASH_REGEX = /\/([\da-f]{40})$/i

export const resolveSnippetFileInfo = (codeView: HTMLElement): FileInfo => {
    // A snippet code view contains a link to the snippet's commit.
    // We use it to find the 40-character commit id.
    const commitLinkElement = codeView.querySelector('a.commit-tease-sha') as HTMLAnchorElement
    if (!commitLinkElement) {
        throw new Error('Could not find commit link in snippet code view')
    }
    const commitIDMatch = commitLinkElement.href.match(COMMIT_HASH_REGEX)
    if (!commitIDMatch?.[1]) {
        throw new Error(`Could not parse commitID from snippet commit link href: ${commitLinkElement.href}`)
    }
    const commitID = commitIDMatch[1]

    // We then use the permalink to determine the repo name and parse the filePath.
    const selector = 'a:not(.commit-tease-sha)'
    const anchors = codeView.querySelectorAll(selector)
    const snippetPermalinkURL = new URL((anchors[0] as HTMLAnchorElement).href)
    const parsedURL = parseURL(snippetPermalinkURL)
    if (parsedURL.pageType !== 'blob') {
        throw new Error(`Snippet URL does not match a blob url: ${snippetPermalinkURL.href}`)
    }
    const { revAndFilePath, rawRepoName } = parsedURL
    if (!revAndFilePath.startsWith(commitID)) {
        throw new Error(
            `Could not parse filePath: revAndFilePath ${revAndFilePath} does not start with commitID ${commitID}`
        )
    }
    const filePath = revAndFilePath.slice(commitID.length + 1)
    return {
        rawRepoName,
        filePath,
        commitID,
        rev: commitID,
    }
}
