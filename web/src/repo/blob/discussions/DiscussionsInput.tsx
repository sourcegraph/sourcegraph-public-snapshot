import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import reactive from 'rx-component'
import { merge, Observable, of, Subject } from 'rxjs'
import { catchError, concat, filter, map, mergeMap, scan, startWith, tap, withLatestFrom } from 'rxjs/operators'
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
    submitLabel: string
    noTitle?: boolean

    previewLoading?: boolean
    previewHTML?: string
}

type Update = (s: State) => State

export const DiscussionsInput = reactive<Props>(props => {
    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const titleInputChanges = new Subject<string>()
    const nextTitleInputChange = (e: React.ChangeEvent<HTMLInputElement>) =>
        titleInputChanges.next(e.currentTarget.value)

    const textAreaKeyDowns = new Subject<React.KeyboardEvent<HTMLTextAreaElement>>()
    const nextTextAreaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => textAreaKeyDowns.next(e)

    const textAreaChanges = new Subject<string>()
    const nextTextAreaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) =>
        textAreaChanges.next(e.currentTarget.value)

    const tabChanges = new Subject<string>()
    const nextTabChange = (tab: string) => tabChanges.next(tab)

    // TODO(slimsag:discussions): ASAP: "preview" tab does not get reset after you submit a comment

    return merge(
        props.pipe(
            map(
                ({ submitLabel, noTitle }): Update => state => ({
                    ...state,
                    submitLabel,
                    noTitle,
                })
            )
        ),

        titleInputChanges.pipe(map((titleInputValue): Update => state => ({ ...state, titleInputValue }))),
        textAreaChanges.pipe(map((textAreaValue): Update => state => ({ ...state, textAreaValue }))),

        // Combine form submits and keyboard shortcut submits
        merge(
            submits.pipe(tap(e => e.preventDefault())),

            // cmd+enter (darwin) or ctrl+enter (linux/win)
            textAreaKeyDowns.pipe(filter(e => (e.ctrlKey || e.metaKey) && e.key === 'Enter'))
        ).pipe(
            tap(e => eventLogger.log('RepliedToDiscussion')), // TODO(slimsag:discussions): not the right event for creating a thread
            withLatestFrom(textAreaChanges, titleInputChanges.pipe(startWith('')), props),
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
        ),

        // Handle tab changes by logging the event and fetching preview data.
        tabChanges.pipe(
            tap(tab => {
                if (tab === 'write') {
                    eventLogger.log('DiscussionsInputWriteTabSelected')
                } else if (tab === 'preview') {
                    eventLogger.log('DiscussionsInputPreviewTabSelected')
                }
            }),
            filter(tab => tab === 'preview'),
            withLatestFrom(textAreaChanges),
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
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(
            ({
                submitLabel,
                noTitle,
                titleInputValue,
                textAreaValue,
                submitting,
                previewLoading,
                previewHTML,
                error,
            }: State): JSX.Element | null => (
                <Form className="discussions-input" onSubmit={nextSubmit}>
                    {!noTitle && (
                        <input
                            className="form-control discussions-input__title"
                            placeholder="Title"
                            autoFocus={true}
                            onChange={nextTitleInputChange}
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
                        onSelectTab={nextTabChange}
                    >
                        <div key="write">
                            <textarea
                                className="form-control discussions-input__text-box"
                                placeholder="Leave a comment"
                                onChange={nextTextAreaChange}
                                onKeyDown={nextTextAreaKeyDown}
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
                            {submitLabel}
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
        )
    )
})

// TODO(slimsag:discussions): ASAP: "Error posting comment" should have different message for thread creation
