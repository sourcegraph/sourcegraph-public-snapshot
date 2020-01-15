import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { uniqueId } from 'lodash'
import * as React from 'react'
import { concat, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mergeMap, startWith, switchMap, tap, withLatestFrom } from 'rxjs/operators'
import { CodeEditorData, EditorId } from '../../../../../shared/src/api/client/services/editorService'
import { TextModel } from '../../../../../shared/src/api/client/services/modelService'
import { COMMENT_URI_SCHEME } from '../../../../../shared/src/api/client/types/textDocument'
import { EditorTextField } from '../../../../../shared/src/components/editorTextField/EditorTextField'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import {
    Spacer,
    TabBorderClassName,
    TabsWithLocalStorageViewStatePersistence,
} from '../../../../../shared/src/components/Tabs'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError } from '../../../../../shared/src/util/errors'
import { Form } from '../../../components/Form'
import { WebEditorCompletionWidget } from '../../../components/shared'
import { renderMarkdown } from '../../../discussions/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'

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

interface Props extends ExtensionsControllerProps {
    location: H.Location
    history: H.History

    /** The initial contents (used when editing an existing comment). */
    initialContents?: string

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
    editorId?: string
    modelUri?: string
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
    private nextSubmit = (e: React.FormEvent<HTMLFormElement>): void => that.submits.next(e)

    private titleInputChanges = new Subject<string>()
    private nextTitleInputChange = (e: React.ChangeEvent<HTMLInputElement>): void =>
        that.titleInputChanges.next(e.currentTarget.value)

    private textAreaKeyDowns = new Subject<React.KeyboardEvent<HTMLTextAreaElement>>()
    private nextTextAreaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
        that.textAreaKeyDowns.next(e)
    }

    private valueChanges = new Subject<string>()
    private nextTextAreaChange = (value: string): void => {
        that.valueChanges.next(value)
    }

    private tabChanges = new Subject<string>()
    private nextTabChange = (tab: string): void => that.tabChanges.next(tab)

    private textAreaRef = React.createRef<HTMLTextAreaElement>()

    public state: State = {
        titleInputValue: '',
        textAreaValue: '',
        submitting: false,
    }

    // TODO(slimsag:discussions): ASAP: "preview" tab does not get reset after you submit a comment

    public componentDidMount(): void {
        const textAreaValueChanges = that.valueChanges.pipe(startWith(that.props.initialContents || ''))

        // Update input model and editor.
        const editorResets = new Subject<void>()
        that.subscriptions.add(
            editorResets.subscribe(() => {
                that.setState({ editorId: undefined, modelUri: undefined })
            })
        )
        const editorInstantiations = editorResets.pipe(
            startWith(undefined),
            switchMap(
                () =>
                    new Observable<EditorId & { modelUri: CodeEditorData['resource'] }>(sub => {
                        const model: TextModel = {
                            uri: uniqueId(`${COMMENT_URI_SCHEME}://`),
                            languageId: 'plaintext',
                            text: that.props.initialContents || '',
                        }
                        that.props.extensionsController.services.model.addModel(model)
                        const editor = that.props.extensionsController.services.editor.addEditor({
                            type: 'CodeEditor',
                            resource: model.uri,
                            selections: [],
                            isActive: true,
                        })
                        sub.next({ editorId: editor.editorId, modelUri: model.uri })
                        return () => {
                            that.props.extensionsController.services.editor.removeEditor(editor)
                        }
                    })
            )
        )

        that.subscriptions.add(
            merge(
                that.titleInputChanges.pipe(
                    tap(titleInputValue => that.props.onTitleChange && that.props.onTitleChange(titleInputValue)),
                    map((titleInputValue): Update => state => ({ ...state, titleInputValue }))
                ),

                textAreaValueChanges.pipe(
                    map(
                        (textAreaValue): Update => state => {
                            if (that.props.titleMode === TitleMode.Implicit) {
                                that.titleInputChanges.next(textAreaValue.trimLeft().split('\n')[0])
                            }
                            return { ...state, textAreaValue }
                        }
                    )
                ),

                editorInstantiations.pipe(
                    map(({ editorId, modelUri }): Update => state => ({ ...state, editorId, modelUri }))
                ),

                // Handle tab changes by logging the event and fetching preview data.
                that.tabChanges.pipe(
                    tap(tab => {
                        if (tab === 'write') {
                            eventLogger.log('DiscussionsInputWriteTabSelected')
                        } else if (tab === 'preview') {
                            eventLogger.log('DiscussionsInputPreviewTabSelected')
                        }
                    }),
                    filter(tab => tab === 'preview'),
                    withLatestFrom(that.valueChanges),
                    mergeMap(([, textAreaValue]) =>
                        concat(
                            of<Update>(state => ({ ...state, previewHTML: undefined, previewLoading: true })),
                            renderMarkdown({ markdown: that.trimImplicitTitle(textAreaValue) }).pipe(
                                map(
                                    (previewHTML): Update => state => ({
                                        ...state,
                                        previewHTML,
                                        previewLoading: false,
                                    })
                                ),
                                catchError((error): Update[] => {
                                    console.error(error)
                                    return [
                                        state => ({
                                            ...state,
                                            error: new Error('Error rendering Markdown: ' + error.message),
                                            previewLoading: false,
                                        }),
                                    ]
                                })
                            )
                        )
                    )
                ),

                // Combine form submits and keyboard shortcut submits
                merge(
                    that.submits.pipe(tap(e => e.preventDefault())),

                    // cmd+enter (darwin) or ctrl+enter (linux/win)
                    that.textAreaKeyDowns.pipe(
                        filter(e => (e.ctrlKey || e.metaKey) && e.key === 'Enter' && that.canSubmit())
                    )
                ).pipe(
                    withLatestFrom(
                        that.valueChanges,
                        that.titleInputChanges.pipe(startWith('')),
                        that.componentUpdates.pipe(startWith(that.props))
                    ),
                    mergeMap(([, textAreaValue, titleInputValue, props]) =>
                        concat(
                            // Start with setting submitting: true
                            of<Update>(state => ({ ...state, submitting: true })),
                            props.onSubmit(titleInputValue, that.trimImplicitTitle(textAreaValue)).pipe(
                                map(
                                    (): Update => state => ({
                                        ...state,
                                        submitting: false,
                                        titleInputValue: '',
                                        textAreaValue: '',
                                    })
                                ),
                                tap(() => editorResets.next()),
                                catchError((error): Update[] => {
                                    console.error(error)
                                    return [
                                        state => ({
                                            ...state,
                                            error: asError(error),
                                            submitting: false,
                                        }),
                                    ]
                                })
                            )
                        )
                    )
                )
            ).subscribe(
                updateState => that.setState(state => updateState(state)),
                err => console.error(err)
            )
        )
        that.componentUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { titleInputValue, editorId, modelUri, error, previewLoading, previewHTML } = that.state

        if (!editorId || !modelUri) {
            return null
        }

        return (
            <Form className="discussions-input" onSubmit={that.nextSubmit}>
                {that.props.titleMode === TitleMode.Explicit && (
                    <input
                        className="form-control discussions-input__title"
                        placeholder="Title"
                        autoFocus={true}
                        onChange={that.nextTitleInputChange}
                        value={titleInputValue}
                    />
                )}
                {/* TODO(slimsag:discussions): local storage persistence is not ideal here. */}
                <TabsWithLocalStorageViewStatePersistence
                    tabs={[
                        { id: 'write', label: 'Write' },
                        { id: 'preview', label: 'Preview' },
                    ]}
                    storageKey="discussions-input-last-tab"
                    tabBarEndFragment={
                        <>
                            <Spacer />
                            <small className={TabBorderClassName}>Markdown supported.</small>
                        </>
                    }
                    tabClassName="tab-bar__tab--h5like"
                    onSelectTab={that.nextTabChange}
                >
                    <div key="write">
                        {that.textAreaRef.current && (
                            <WebEditorCompletionWidget
                                textArea={that.textAreaRef.current}
                                editorId={editorId}
                                extensionsController={that.props.extensionsController}
                            />
                        )}
                        <EditorTextField
                            className="form-control discussions-input__text-box"
                            placeholder="Leave a comment"
                            editorId={editorId}
                            modelUri={modelUri}
                            onValueChange={that.nextTextAreaChange}
                            onKeyDown={that.nextTextAreaKeyDown}
                            textAreaRef={that.textAreaRef}
                            autoFocus={that.props.titleMode !== TitleMode.Explicit}
                            extensionsController={that.props.extensionsController}
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
                        disabled={!that.canSubmit()}
                    >
                        {that.props.submitLabel}
                    </button>
                </div>
                {error && <ErrorAlert className="discussions-input__error" error={error} />}
            </Form>
        )
    }

    /** Trims the implicit title string out of the comment (e.g. textarea value). */
    private trimImplicitTitle = (comment: string): string => {
        if (that.props.titleMode !== TitleMode.Implicit) {
            return comment
        }
        return comment
            .trimLeft()
            .split('\n')
            .slice(1)
            .join('\n')
    }

    private canSubmit = (): boolean => {
        const textAreaEmpty = !that.state.textAreaValue.trim()
        const titleRequired = that.props.titleMode !== TitleMode.None
        const titleEmpty = !that.state.titleInputValue.trim()
        return !that.state.submitting && !textAreaEmpty && (!titleRequired || !titleEmpty)
    }
}
