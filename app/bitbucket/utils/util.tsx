/**
 * BitbucketState contains a dynamic set of properties based on the current page.
 */
export interface BitbucketState {
    commit: BitbucketCommit | false
    currentUser: BitbucketUser | false
    filePath: BitbucketFile | false
    project: BitbucketProject | false
    pullRequest: BitbucketPullRequest | false
    ref: BitbucketRef | false
    repository: BitbucketRepository | false
}

interface BitbucketStateHandler {
    /**
     * Available on pages that show a commit (currently just /projects/PROJ/repos/REPO/commits/HASH).
     * @return {BitbucketCommit} information about the commit.
     */
    getCommit: () => BitbucketCommit

    /**
     * Available on every page.
     * @return {BitbucketUser} information about the current user, if there is a logged in user.
     */
    getCurrentUser: () => BitbucketUser

    /**
     * Available on pages that show a file (Pull Request, Commit, File Browser).
     * @return {BitbucketFile} information about the currently viewed file.
     */
    getFilePath: () => BitbucketFile

    /**
     * Available on every page.
     * @return {BitbucketProject} information about the currently viewed project, if there is one being viewed.
     */
    getProject: () => BitbucketProject

    /**
     * Available on every page.
     * @return {BitbucketPullRequest} information about the currently viewed pull request, if there is one being viewed.
     */
    getPullRequest: () => BitbucketPullRequest

    /**
     * Available on every page.
     * @return {BitbucketRef} information about the currently viewed ref (e.g. currently selected branch), if there is one being viewed.
     */
    getRef: () => BitbucketRef

    /**
     * Available on every page.
     * @return {BitbucketRepository} information about the currently viewed repository, if there is one being viewed.
     */
    getRepository: () => BitbucketRepository
}

export interface BitbucketCommit {
    author: BitbucketParticipant
    authorTimestamp: number
    committer: BitbucketParticipant
    committerTimestamp: number
    displayId: string
    id: string
    message: string
    parents?: BitbucketCommit[]
}

interface BitbucketParticipant {
    avatarUrl: string
    emailAddress: string
    name: string
}

/**
 * Describes a user in Bitbucket Server & Stash.
 */
interface BitbucketUser {
    active: boolean
    avatarUrl?: string
    displayName: string
    emailAddress: string
    id: number
    name: string
    slug: string
    type: string
}

export interface BitbucketProject {
    id: number
    name: string
    key: string
    public: boolean
    avatarUrl: string
}

interface BitbucketFile {
    components: string[]
    extension: string
    name: string
}

interface BitbucketPullRequest {
    author: BitbucketParticipant
    createdDate: string
    description: string
    descriptionAsHtml: string
    id: number
    fromRef: BitbucketRef
    participants: BitbucketParticipant[]
    reviewers: BitbucketParticipant[]
    title: string
    toRef: BitbucketRef
    updatedDate: string
    version: number
}

interface BitbucketRef {
    displayId: string
    id: string
    isDefault: boolean
    hash: string
    latestCommit: string
    repository: BitbucketRepository
    type: {
        id: 'tag' | 'branch' | 'commit'
        name: 'Tag' | 'Branch' | 'Commit'
    }
}

export interface BitbucketRepository {
    id: number
    name: string
    slug: string
    project: BitbucketProject
    scmId: string
    public: boolean
    cloneUrl: string
}

const BITBUCKET_STATE_ELEMENT_ID = 'BITBUCKET_STATE_ID'
const BITBUCKET_LINE_SCROLL_ID = 'BITBUCKET_LINE_SCROLL_ID'

export function configureBitbucketHandlers(): void {
    bitbucketPierce(getBitbucketStateHandler, BITBUCKET_STATE_ELEMENT_ID)
}

function getBitbucketStateHandler(): void {
    window.require(['bitbucket/util/state'], (bitbucketStateHandler: BitbucketStateHandler) => {
        document.dispatchEvent(
            new CustomEvent('bitbucketLoaded', {
                detail: {
                    commit: bitbucketStateHandler.getCommit(),
                    pullRequest: bitbucketStateHandler.getPullRequest(),
                    project: bitbucketStateHandler.getProject(),
                    currentUser: bitbucketStateHandler.getCurrentUser(),
                    filePath: bitbucketStateHandler.getFilePath(),
                    ref: bitbucketStateHandler.getRef(),
                    repository: bitbucketStateHandler.getRepository(),
                } as BitbucketState,
            })
        )
    })
}

export function scrollToLine(lineNumber: number): void {
    let s = document.getElementById(BITBUCKET_LINE_SCROLL_ID) as HTMLScriptElement
    if (s) {
        s.remove()
    }
    s = document.createElement('script') as HTMLScriptElement
    s.id = BITBUCKET_LINE_SCROLL_ID
    s.setAttribute('type', 'text/javascript')
    const scrollFunc = `function handleScroll() {
        var editor = AJS.$('.CodeMirror')[0].CodeMirror;
        window.scrollTo(null, editor.heightAtLine(${lineNumber}) - 100);
    }`
    s.textContent = scrollFunc + ';' + `handleScroll();`
    document.body.appendChild(s)
}

/**
 * This injects code as a script tag into a web page body.
 * Needed to reference the Bitbucket Internal AJS code.
 */
function bitbucketPierce(code: () => void, id: string): void {
    let s = document.getElementById(id) as HTMLScriptElement
    if (s) {
        return
    }
    s = document.createElement('script') as HTMLScriptElement
    s.id = id
    s.setAttribute('type', 'text/javascript')
    s.textContent = code.toString() + ';' + code.name + '();'
    document.body.appendChild(s)
}

export function getRevisionState(state: BitbucketState): { headRev: string; baseRev?: string } | undefined {
    const { commit, ref, pullRequest, repository } = state
    if (!repository) {
        return undefined
    }
    let baseRev: string | undefined
    let headRev: string | undefined
    if (commit) {
        headRev = commit.id
    } else if (ref) {
        headRev = ref.latestCommit
    } else if (pullRequest) {
        headRev = pullRequest.fromRef.latestCommit
        baseRev = pullRequest.toRef.latestCommit
    }

    // Parse the URL to see if it contains the previous commit info.
    const query = new URLSearchParams(window.location.search)
    if (query.has('until')) {
        headRev = query.get('until')!
    } else if (query.has('at')) {
        headRev = query.get('at')!
    }
    if (query.has('since')) {
        baseRev = query.get('since')!
    }
    if (!headRev) {
        return undefined
    }
    return { headRev, baseRev }
}
