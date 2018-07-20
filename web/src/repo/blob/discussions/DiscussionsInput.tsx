import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, concat, filter, map, mergeMap, startWith, tap, withLatestFrom } from 'rxjs/operators'
import { Form } from '../../../components/Form'
import { Markdown } from '../../../components/Markdown'
import { Spacer, TabBorderClassName, TabsWithLocalStorageViewStatePersistence } from '../../../components/Tabs'
import { eventLogger } from '../../../tracking/eventLogger'
import { renderMarkdown } from './DiscussionsBackend'

interface Props {
    /** The label to display on the submit button. */
    submitLabel: string

    /** Called when the submit button is clicked. */
    onSubmit: (title: string, textAreaValue: string) => Observable<void>

    /** Whether or not to hide the title portion of the input. */
    noTitle?: boolean
}

interface State {
    titleInputValue: string
    textAreaValue: string
    submitting: boolean
    error?: any

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
    private nextTextAreaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => this.textAreaKeyDowns.next(e)

    private textAreaChanges = new Subject<string>()
    private nextTextAreaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) =>
        this.textAreaChanges.next(e.currentTarget.value)

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
                    tap(x => console.log(x)),
                    map((titleInputValue): Update => state => ({ ...state, titleInputValue }))
                ),

                this.textAreaChanges.pipe(map((textAreaValue): Update => state => ({ ...state, textAreaValue }))),

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
                    mergeMap(([, textAreaValue]) =>
                        of<Update>(state => ({ ...state, previewHTML: undefined, previewLoading: true })).pipe(
                            concat(
                                renderMarkdown(textAreaValue).pipe(
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
                                            return [state => ({ ...state, error, previewLoading: false })]
                                        }
                                    )
                                )
                            )
                        )
                    )
                ),

                // Combine form submits and keyboard shortcut submits
                merge(
                    this.submits.pipe(tap(e => e.preventDefault())),

                    // cmd+enter (darwin) or ctrl+enter (linux/win)
                    this.textAreaKeyDowns.pipe(filter(e => (e.ctrlKey || e.metaKey) && e.key === 'Enter'))
                ).pipe(
                    tap(e => eventLogger.log('RepliedToDiscussion')), // TODO(slimsag:discussions): not the right event for creating a thread
                    withLatestFrom(
                        this.textAreaChanges,
                        this.titleInputChanges.pipe(startWith('')),
                        this.componentUpdates.pipe(startWith(this.props))
                    ),
                    mergeMap(([, textAreaValue, titleInputValue, props]) =>
                        // Start with setting submitting: true
                        of<Update>(state => ({ ...state, submitting: true })).pipe(
                            concat(
                                props.onSubmit(titleInputValue, textAreaValue).pipe(
                                    map(
                                        (): Update => state => ({
                                            ...state,
                                            submitting: false,
                                            titleInputValue: '',
                                            textAreaValue: '',
                                        })
                                    ),
                                    catchError(
                                        (error): Update[] => {
                                            console.error(error)
                                            return [state => ({ ...state, error, submitting: false })]
                                        }
                                    )
                                )
                            )
                        )
                    )
                )
            ).subscribe(updateState => this.setState(state => updateState(state)), err => console.error(err))
        )
    }

    public componentWillReceiveProps(props: Props): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { titleInputValue, textAreaValue, submitting, error, previewLoading, previewHTML } = this.state

        return (
            <Form className="discussions-input" onSubmit={this.nextSubmit}>
                {!this.props.noTitle && (
                    <input
                        className="form-control discussions-input__title"
                        placeholder="Title"
                        autoFocus={true}
                        onChange={this.nextTitleInputChange}
                        value={titleInputValue}
                    />
                )}
                {/* TODO(slimsag:discussions): ASAP: local storage persistence is not ideal here. */}
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
                        />
                    </div>
                    <div key="preview" className="discussions-input__preview">
                        {previewLoading && <LoaderIcon className="icon-inline" />}
                        {!previewLoading && previewHTML && <Markdown dangerousInnerHTML={previewHTML} />}
                    </div>
                </TabsWithLocalStorageViewStatePersistence>
                <div className="discussions-input__row">
                    <button
                        type="submit"
                        className="btn btn-primary discussions-input__button"
                        disabled={submitting || !textAreaValue}
                    >
                        {this.props.submitLabel}
                    </button>
                </div>
                {error && (
                    <div className="discussions-input__error alert alert-danger">
                        <ErrorIcon className="icon-inline discussions-input__error-icon" />
                        Error posting comment: {error.message}
                    </div>
                )}
            </Form>
        )
    }
}

// TODO(slimsag:discussions): ASAP: "Error posting comment" should have different message for thread creation
