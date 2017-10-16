import ChatIcon from '@sourcegraph/icons/lib/Chat'
import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as H from 'history'
import * as React from 'react'
import { match } from 'react-router'
import reactive from 'rx-component'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { RepoNav } from '../repo/RepoNav'
import { toEditorURL } from '../util/url'
import { fetchSharedItem } from './backend'
import { Comment } from './Comment'

const SharedItemNotFound = () => <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, we can&#39;t find anything here.' />

interface Props {
    match: match<{ ulid: string }>
    location: H.Location
    history: H.History
}

interface State {
    sharedItem?: GQL.ISharedItem
    location: H.Location
    history: H.History
}

type Update = (s: State) => State

/**
 * Renders a shared code comment's thread.
 */
export const CommentsPage = reactive<Props>(props =>
    props
        .mergeMap(props =>
            fetchSharedItem(props.match.params.ulid)
                .map(sharedItem => (state: State): State => ({ ...state, location: props.location, history: props.history, sharedItem: sharedItem || undefined }))
        )
        .scan<Update, State>((state: State, update: Update) => update(state), undefined)
        .map(({ location, history, sharedItem }: State): JSX.Element | null => {
            if (!sharedItem) {
                return <SharedItemNotFound />
            }

            const editorURL = toEditorURL(
                sharedItem.thread.repo.remoteUri,
                sharedItem.thread.revision,
                sharedItem.thread.file,
                {line: sharedItem.thread.startLine},
                sharedItem.thread.id
            )
            const openEditor = () => {
                window.open(editorURL, 'sourcegraph-editor')
            }

            return (
                <div className='comments-page'>
                    <PageTitle title={getPageTitle(sharedItem)} />
                    {/* TODO(slimsag): future: do not disable breadcrumb _if_ the repository is public */}
                    <RepoNav
                        repoPath={sharedItem.thread.repo.remoteUri}
                        rev={sharedItem.thread.branch || sharedItem.thread.repoRevision}
                        filePath={sharedItem.thread.file}
                        hideCopyLink={true}
                        showOpenOnDesktop={true}
                        customEditorURL={editorURL}
                        breadcrumbDisabled={true}
                        revSwitcherDisabled={true}
                        location={location}
                        history={history}
                    />
                    <div className='comments-page__content'>
                        {sharedItem && !sharedItem.thread.lines && <div className='comments-page__no-shared-code-container'>
                            <div className='comments-page__no-shared-code'>
                                The author of this discussion did not <a href='https://about.sourcegraph.com/docs/editor/share-code'>share the code</a>.&nbsp;
                                <a href='' onClick={openEditor}>Open in Sourcegraph Editor</a> to see code.
                            </div>
                        </div>}
                        {sharedItem && codeViewComponent(sharedItem)}
                        <hr className='comments-page__hr' />
                        {sharedItem && sharedItem.thread.comments.map(comment =>
                            <div className='comments-page__comment-container' key={comment.id} id={String(comment.id)}>
                                <Comment location={location} comment={comment} />
                                <hr className='comments-page__hr' />
                            </div>
                        )}
                        <button className='btn btn-primary btn-block comments-page__reply-in-editor' onClick={openEditor}>
                            Reply in Sourcegraph Editor
                        </button>
                    </div>
                </div>
            )
        })
        .catch(err => {
            console.error(err)
            return []
        })
)

function getPageTitle(sharedItem: GQL.ISharedItem): string | undefined {
    const title = sharedItem.comment ? sharedItem.comment.title : sharedItem.thread.title
    if (title === '') {
        // TODO(slimsag): future: Maybe serve some other information here. It
        // can happen for e.g. a code snippet ('thread') without any comments
        // on it.
        return undefined // "Sourcegraph"
    }
    return title
}

interface Line {
    number: number
    content: string
    isStartLine: boolean
    className: string
}

const phonyBeforeLines = [
    'func (r *commitResolver) File(ctx context.Context, args *struct {',
    '	Path string',
    '}) (*fileResolver, error) {',
]

const phonyLines = [
    '	return &fileResolver{',
    '		commit: r.commit,',
    '		name:   path.Base(args.Path),',
]

const phonyAfterLines = [
    '		path:   args.Path,',
    '	}, nil',
    '}',
]

const itemToLines = (sharedItem: GQL.ISharedItem): Line[] => {
    const startLine = sharedItem.thread.startLine
    const threadLines = sharedItem.thread.lines
    const htmlBefore = threadLines ? threadLines.htmlBefore.split('\n') : phonyBeforeLines
    const html = threadLines ? threadLines.html.split('\n') : phonyLines
    const htmlAfter = threadLines ? threadLines.htmlAfter.split('\n') : phonyAfterLines
    let lines = htmlBefore.map((line: string, i: number) => ({
        number: startLine - (htmlBefore.length - i),
        content: line,
        isStartLine: false,
        className: 'comments-page__line--before',
    }))
    lines = lines.concat(html.map((line: string, i: number) => ({
        number: startLine + i,
        content: line,
        isStartLine: false,
        className: 'comments-page__line--main',
    })))
    lines = lines.concat(htmlAfter.map((line: string, i: number) => ({
        number: startLine + i + html.length,
        content: line,
        isStartLine: false,
        className: 'comments-page__line--after',
    })))
    return lines.map((line: Line) => ({
        ...line,
        isStartLine: line.number === startLine,
        className: `comments-page__line ${line.className}`,
    }))
}

function codeViewComponent(sharedItem: GQL.ISharedItem): JSX.Element | null {
    return (
        <table className={`comments-page__table${sharedItem.thread.lines ? '' : ' comments-page__table--blurry'}`}>
            <tbody>
                {itemToLines(sharedItem).map((line: Line) => <tr className={line.className} key={line.number}>
                        <td className={`comments-page__line-number${line.isStartLine ? ' comments-page__line-number--start-line' : ''}`}>
                            {line.isStartLine && <ChatIcon />}
                            {line.number}
                        </td>
                        <td className='comments-page__line-content'><pre className='comments-page__pre' dangerouslySetInnerHTML={{__html: line.content}}></pre></td>
                    </tr>
                )}
            </tbody>
        </table>
    )
}
