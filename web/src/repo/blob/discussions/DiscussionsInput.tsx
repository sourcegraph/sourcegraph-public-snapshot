import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { concat, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mergeMap, startWith, tap, withLatestFrom } from 'rxjs/operators'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import {
    Spacer,
    TabBorderClassName,
    TabsWithLocalStorageViewStatePersistence,
} from '../../../../../shared/src/components/Tabs'
import { asError } from '../../../../../shared/src/util/errors'
import { Form } from '../../../components/Form'
import { renderMarkdown } from '../../../discussions/backend'
import { eventLogger } from '../../../tracking/eventLogger'

/**
 * How & whether or not to render a title input field.
 */
export enum TitleMode {
    /** Explicitly show a separate title input field. */
    Explicit,

    /** Implicitly use the first line of the main textarea as the title field (like Git commit messages). */
    Implicit,

    /** No title input at all, e.g. for replying to discussion threads.  */
    None,
}

interface Props {
    location: H.Location
    history: H.History

    /** The label to display on the submit button. */
    submitLabel: string

    /** Called when the submit button is clicked. */
    onSubmit: (title: string, comment: string) => Observable<void>

    /** How & whether or not to render a title input field. */
    titleMode: TitleMode

    /** Called when the title value changes. */
    onTitleChange?: (title: string) => void
}

interface State {
    titleInputValue: string
    textAreaValue: string
    submitting: boolean
    error?: Error

    previewLoading?: boolean
    previewHTML?: string
}

type Update = (s: State) => State

export class DiscussionsInput extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private nextSubmit = (e: React.FormEvent<HTMLFormElement>) => this.submits.next(e)

    private titleInputChanges = new Subject<string>()
    private nextTitleInputChange = (e: React.ChangeEvent<HTMLInputElement>) =>
        this.titleInputChanges.next(e.currentTarget.value)

    private textAreaKeyDowns = new Subject<React.KeyboardEvent<HTMLTextAreaElement>>()
    private nextTextAreaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        this.textAreaKeyDowns.next(e)
    }

    private textAreaChanges = new Subject<{ textAreaValue: string; selectionStart: number; element?: HTMLElement }>()
    private nextTextAreaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        this.textAreaChanges.next({
            textAreaValue: e.currentTarget.value,
            selectionStart: e.currentTarget.selectionStart,
            element: e.currentTarget,
        })
    }

    private tabChanges = new Subject<string>()
    private nextTabChange = (tab: string) => this.tabChanges.next(tab)

    public state: State = {
        titleInputValue: '',
        textAreaValue: '',
        submitting: false,
    }

    // TODO(slimsag:discussions): ASAP: "preview" tab does not get reset after you submit a comment

    public componentDidMount(): void {
        this.subscriptions.add(
            merge(
                this.titleInputChanges.pipe(
                    tap(titleInputValue => this.props.onTitleChange && this.props.onTitleChange(titleInputValue)),
                    map((titleInputValue): Update => state => ({ ...state, titleInputValue }))
                ),

                this.textAreaChanges.pipe(
                    startWith({ textAreaValue: '', selectionStart: 0, element: undefined }),
                    map(
                        (textArea): Update => state => {
                            if (this.props.titleMode === TitleMode.Implicit) {
                                this.titleInputChanges.next(textArea.textAreaValue.trimLeft().split('\n')[0])
                            }
                            return { ...state, textArea }
                        }
                    )
                ),

                // Handle tab changes by logging the event and fetching preview data.
                this.tabChanges.pipe(
                    tap(tab => {
                        if (tab === 'write') {
                            eventLogger.log('DiscussionsInputWriteTabSelected')
                        } else if (tab === 'preview') {
                            eventLogger.log('DiscussionsInputPreviewTabSelected')
                        }
                    }),
                    filter(tab => tab === 'preview'),
                    withLatestFrom(this.textAreaChanges),
                    mergeMap(([, { textAreaValue }]) =>
                        concat(
                            of<Update>(state => ({ ...state, previewHTML: undefined, previewLoading: true })),
                            renderMarkdown({ markdown: this.trimImplicitTitle(textAreaValue) }).pipe(
                                map(
                                    (previewHTML): Update => state => ({
                                        ...state,
                                        previewHTML,
                                        previewLoading: false,
                                    })
                                ),
                                catchError(
                                    (error): Update[] => {
                                        console.error(error)
                                        return [
                                            state => ({
                                                ...state,
                                                error: new Error('Error rendering Markdown: ' + error.message),
                                                previewLoading: false,
                                            }),
                                        ]
                                    }
                                )
                            )
                        )
                    )
                ),

                // Combine form submits and keyboard shortcut submits
                merge(
                    this.submits.pipe(tap(e => e.preventDefault())),

                    // cmd+enter (darwin) or ctrl+enter (linux/win)
                    this.textAreaKeyDowns.pipe(
                        filter(e => (e.ctrlKey || e.metaKey) && e.key === 'Enter' && this.canSubmit())
                    )
                ).pipe(
                    withLatestFrom(
                        this.textAreaChanges,
                        this.titleInputChanges.pipe(startWith('')),
                        this.componentUpdates.pipe(startWith(this.props))
                    ),
                    mergeMap(([, { textAreaValue }, titleInputValue, props]) =>
                        concat(
                            // Start with setting submitting: true
                            of<Update>(state => ({ ...state, submitting: true })),
                            props.onSubmit(titleInputValue, this.trimImplicitTitle(textAreaValue)).pipe(
                                map(
                                    (): Update => state => ({
                                        ...state,
                                        submitting: false,
                                        titleInputValue: '',
                                        textArea: { ...state, textAreaValue: '', selectionStart: 0 },
                                    })
                                ),
                                catchError(
                                    (error): Update[] => {
                                        console.error(error)
                                        return [
                                            state => ({
                                                ...state,
                                                error: asError(error),
                                                submitting: false,
                                            }),
                                        ]
                                    }
                                )
                            )
                        )
                    )
                )
            ).subscribe(updateState => this.setState(state => updateState(state)), err => console.error(err))
        )
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { titleInputValue, textAreaValue, error, previewLoading, previewHTML } = this.state

        return (
            <Form className="discussions-input" onSubmit={this.nextSubmit}>
                {this.props.titleMode === TitleMode.Explicit && (
                    <input
                        className="form-control discussions-input__title"
                        placeholder="Title"
                        autoFocus={true}
                        onChange={this.nextTitleInputChange}
                        value={titleInputValue}
                    />
                )}
                {/* TODO(slimsag:discussions): local storage persistence is not ideal here. */}
                <TabsWithLocalStorageViewStatePersistence
                    tabs={[{ id: 'write', label: 'Write' }, { id: 'preview', label: 'Preview' }]}
                    storageKey="discussions-input-last-tab"
                    tabBarEndFragment={
                        <>
                            <Spacer />
                            <small className={TabBorderClassName}>Markdown supported.</small>
                        </>
                    }
                    tabClassName="tab-bar__tab--h5like"
                    onSelectTab={this.nextTabChange}
                >
                    <div key="write">
                        <textarea
                            className="form-control discussions-input__text-box"
                            placeholder="Leave a comment"
                            onChange={this.nextTextAreaChange}
                            onKeyDown={this.nextTextAreaKeyDown}
                            value={textAreaValue}
                            autoFocus={this.props.titleMode !== TitleMode.Explicit}
                        />
                    </div>
                    <div key="preview" className="discussions-input__preview">
                        {previewLoading && <LoadingSpinner className="icon-inline" />}
                        {!previewLoading && previewHTML && <Markdown dangerousInnerHTML={previewHTML} />}
                    </div>
                </TabsWithLocalStorageViewStatePersistence>
                <div className="discussions-input__row">
                    <button
                        type="submit"
                        className="btn btn-primary discussions-input__button"
                        disabled={!this.canSubmit()}
                    >
                        {this.props.submitLabel}
                    </button>
                </div>
                {error && (
                    <div className="discussions-input__error alert alert-danger">
                        <AlertCircleIcon className="icon-inline discussions-input__error-icon" />
                        {error.message}
                    </div>
                )}
            </Form>
        )
    }

    /** Trims the implicit title string out of the comment (e.g. textarea value). */
    private trimImplicitTitle = (comment: string): string => {
        if (this.props.titleMode !== TitleMode.Implicit) {
            return comment
        }
        return comment
            .trimLeft()
            .split('\n')
            .slice(1)
            .join('\n')
    }

    private canSubmit = (): boolean => {
        const textAreaEmpty = !this.state.textAreaValue.trim()
        const titleRequired = this.props.titleMode !== TitleMode.None
        const titleEmpty = !this.state.titleInputValue.trim()
        return !this.state.submitting && !textAreaEmpty && (!titleRequired || !titleEmpty)
    }
}
