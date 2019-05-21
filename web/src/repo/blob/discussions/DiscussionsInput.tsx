import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { uniqueId } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { concat, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mergeMap, startWith, switchMap, tap, withLatestFrom } from 'rxjs/operators'
import { CodeEditor, EditorId } from '../../../../../shared/src/api/client/services/editorService'
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

    /**
     * If set, a "Discard" button is shown, and this callback is called when it is clicked.
     */
    onDiscard?: () => void

    className?: string
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
    private nextSubmit = (e: React.FormEvent<HTMLFormElement>) => this.submits.next(e)

    private titleInputChanges = new Subject<string>()
    private nextTitleInputChange = (e: React.ChangeEvent<HTMLInputElement>) =>
        this.titleInputChanges.next(e.currentTarget.value)

    private textAreaKeyDowns = new Subject<React.KeyboardEvent<HTMLTextAreaElement>>()
    private nextTextAreaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        this.textAreaKeyDowns.next(e)
    }

    private valueChanges = new Subject<string>()
    private nextTextAreaChange = (value: string) => {
        this.valueChanges.next(value)
    }

    private tabChanges = new Subject<string>()
    private nextTabChange = (tab: string) => this.tabChanges.next(tab)

    private textAreaRef = React.createRef<HTMLTextAreaElement>()

    public state: State = {
        titleInputValue: '',
        textAreaValue: '',
        submitting: false,
    }

    // TODO(slimsag:discussions): ASAP: "preview" tab does not get reset after you submit a comment

    public componentDidMount(): void {
        const textAreaValueChanges = this.valueChanges.pipe(startWith(this.props.initialContents || ''))

        // Update input model and editor.
        const editorResets = new Subject<void>()
        this.subscriptions.add(
            editorResets.subscribe(() => {
                this.setState({ editorId: undefined, modelUri: undefined })
            })
        )
        const editorInstantiations = editorResets.pipe(
            startWith(void 0),
            switchMap(
                () =>
                    new Observable<EditorId & { modelUri: CodeEditor['resource'] }>(sub => {
                        const model: TextModel = {
                            uri: uniqueId(`${COMMENT_URI_SCHEME}://`),
                            languageId: 'plaintext',
                            text: this.props.initialContents || '',
                        }
                        this.props.extensionsController.services.model.addModel(model)
                        const editor = this.props.extensionsController.services.editor.addEditor({
                            type: 'CodeEditor',
                            resource: model.uri,
                            selections: [],
                            isActive: true,
                        })
                        sub.next({ editorId: editor.editorId, modelUri: model.uri })
                        return () => {
                            this.props.extensionsController.services.editor.removeEditor(editor)
                            this.props.extensionsController.services.model.removeModel(model.uri)
                        }
                    })
            )
        )

        this.subscriptions.add(
            merge(
                this.titleInputChanges.pipe(
                    tap(titleInputValue => this.props.onTitleChange && this.props.onTitleChange(titleInputValue)),
                    map((titleInputValue): Update => state => ({ ...state, titleInputValue }))
                ),

                textAreaValueChanges.pipe(
                    map(
                        (textAreaValue): Update => state => {
                            if (this.props.titleMode === TitleMode.Implicit) {
                                this.titleInputChanges.next(textAreaValue.trimLeft().split('\n')[0])
                            }
                            return { ...state, textAreaValue }
                        }
                    )
                ),

                editorInstantiations.pipe(
                    map(({ editorId, modelUri }): Update => state => ({ ...state, editorId, modelUri }))
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
                    withLatestFrom(this.valueChanges),
                    mergeMap(([, textAreaValue]) =>
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
                    this.submits.pipe(tap(e => e.preventDefault())),

                    // cmd+enter (darwin) or ctrl+enter (linux/win)
                    this.textAreaKeyDowns.pipe(
                        filter(e => (e.ctrlKey || e.metaKey) && e.key === 'Enter' && this.canSubmit())
                    )
                ).pipe(
                    withLatestFrom(
                        this.valueChanges,
                        this.titleInputChanges.pipe(startWith('')),
                        this.componentUpdates.pipe(startWith(this.props))
                    ),
                    mergeMap(([, textAreaValue, titleInputValue, props]) =>
                        concat(
                            // Start with setting submitting: true
                            of<Update>(state => ({ ...state, submitting: true })),
                            props.onSubmit(titleInputValue, this.trimImplicitTitle(textAreaValue)).pipe(
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
        const { titleInputValue, editorId, modelUri, error, previewLoading, previewHTML } = this.state

        if (!editorId || !modelUri) {
            return null
        }

        return (
            <Form className={`discussions-input ${this.props.className || ''}`} onSubmit={this.nextSubmit}>
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
                        {this.textAreaRef.current && (
                            <WebEditorCompletionWidget
                                textArea={this.textAreaRef.current}
                                editorId={editorId}
                                extensionsController={this.props.extensionsController}
                            />
                        )}
                        <EditorTextField
                            className="form-control discussions-input__text-box"
                            placeholder="Leave a comment"
                            editorId={editorId}
                            modelUri={modelUri}
                            onValueChange={this.nextTextAreaChange}
                            onKeyDown={this.nextTextAreaKeyDown}
                            textAreaRef={this.textAreaRef}
                            autoFocus={this.props.titleMode !== TitleMode.Explicit}
                            extensionsController={this.props.extensionsController}
                        />
                    </div>
                    <div key="preview" className="discussions-input__preview">
                        {previewLoading && <LoadingSpinner className="icon-inline" />}
                        {!previewLoading && previewHTML && <Markdown dangerousInnerHTML={previewHTML} />}
                    </div>
                </TabsWithLocalStorageViewStatePersistence>
                <div className="discussions-input__row">
                    {this.props.onDiscard && (
                        <button
                            type="reset"
                            className="btn btn-link discussions-input__button"
                            disabled={this.state.submitting}
                            onClick={this.props.onDiscard}
                        >
                            Discard
                        </button>
                    )}
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
