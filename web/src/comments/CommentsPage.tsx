import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as H from 'history'
import * as React from 'react'
import { match } from 'react-router'
import reactive from 'rx-component'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import { Observable } from 'rxjs/Observable'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { RepoNav } from '../repo/RepoNav'
import { toEditorURL } from '../util/url'
import { fetchSharedItem } from './backend'
import { CodeView } from './CodeView'
import { Comment } from './Comment'

const SharedItemNotFound = () => <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, we can&#39;t find anything here.' />

interface Props {
    match: match<{ ulid: string }>
    location: H.Location
    history: H.History
}

interface State {
    sharedItem?: GQL.ISharedItem | null
    location: H.Location
    history: H.History
    error?: Error
}

type Update = (s: State) => State

/**
 * Renders a shared code comment's thread.
 */
export const CommentsPage = reactive<Props>(props =>
    Observable.merge(
        props
            .map(({ location, history }): Update => state => ({ ...state, location, history })),

        props
            .map(props => props.match.params.ulid)
            .distinctUntilChanged()
            .mergeMap(ulid =>
                fetchSharedItem(ulid)
                    .map((sharedItem): Update => state => ({ ...state, sharedItem }))
                    .catch((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error })]
                    })
            )
    )
        .scan<Update, State>((state: State, update: Update) => update(state), {} as State)
        .map(({ location, history, sharedItem }: State): JSX.Element | null => {
            if (sharedItem === undefined) {
                // TODO(slimsag): future: add loading screen
                return null
            }
            if (sharedItem === null) {
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
                        customEditorURL={editorURL}
                        breadcrumbDisabled={true}
                        revSwitcherDisabled={true}
                        line={sharedItem && sharedItem.thread.startLine}
                        location={location}
                        history={history}
                    />
                    {sharedItem && !sharedItem.thread.repoRevision && <div className='comments-page__no-revision'>
                        <ErrorIcon className='icon-inline comments-page__error-icon'/>
                        {sharedItem.thread.comments.length === 0 ?
                            'This code snippet was created from code that was not pushed. File or line numbers may have changed since this snippet was created.' :
                            'This discussion was created on code that was not pushed. File or line numbers may have changed since this discussion was created.'
                        }
                    </div>}
                    <div className='comments-page__content'>
                        {sharedItem && !sharedItem.thread.lines && <div className='comments-page__no-shared-code-container'>
                            <div className='comments-page__no-shared-code'>
                                The author of this discussion did not <a href='https://about.sourcegraph.com/docs/editor/share-code'>share the code</a>.&nbsp;
                                <a href='' onClick={openEditor}>Open in Sourcegraph Editor</a> to see code.
                            </div>
                        </div>}
                        {sharedItem && CodeView(sharedItem)}
                        {sharedItem && sharedItem.thread.comments.map(comment =>
                            <Comment location={location} comment={comment} key={comment.id} />
                        )}
                        {sharedItem &&
                            <button className='btn btn-primary btn-block comments-page__reply-in-editor' onClick={openEditor}>
                                {sharedItem.thread.comments.length === 0 ? 'Open in Sourcegraph Editor' : 'Reply in Sourcegraph Editor'}
                            </button>
                        }
                    </div>
                </div>
            )
        })
)

function getPageTitle(sharedItem: GQL.ISharedItem): string | undefined {
    const title = sharedItem.comment ? sharedItem.comment.title : sharedItem.thread.title
    if (title === '') {
        return sharedItem.thread.file
    }
    return title
}
