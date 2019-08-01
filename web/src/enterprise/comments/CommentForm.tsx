import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { uniqueId } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { concat, from, merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mergeMap, startWith, switchMap, tap, withLatestFrom } from 'rxjs/operators'
import { CodeEditor, EditorId } from '../../../../shared/src/api/client/services/editorService'
import { TextModel } from '../../../../shared/src/api/client/services/modelService'
import { COMMENT_URI_SCHEME } from '../../../../shared/src/api/client/types/textDocument'
import { EditorTextField } from '../../../../shared/src/components/editorTextField/EditorTextField'
import { Markdown } from '../../../../shared/src/components/Markdown'
import {
    Spacer,
    TabBorderClassName,
    TabsWithLocalStorageViewStatePersistence,
} from '../../../../shared/src/components/Tabs'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { asError } from '../../../../shared/src/util/errors'
import { Form } from '../../components/Form'
import { WebEditorCompletionWidget } from '../../components/shared'
import { renderMarkdown } from '../../discussions/backend'
import { eventLogger } from '../../tracking/eventLogger'

interface Props extends ExtensionsControllerProps {
    // TODO!(sqs): confirm navigation away when field is dirty
    history: H.History

    /** The initial body (used when editing an existing comment). */
    initialBody?: string

    /** The label to display on the submit button. */
    submitLabel: string

    /** Called when the submit button is clicked. */
    onSubmit: (body: string) => Promise<void>

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

// TODO!(sqs): make this support text field completion in extension api
export const CommentForm: React.FunctionComponent<Props> = ({ initialBody }) => {
    const textAreaRef = React.createRef<HTMLTextAreaElement>()

    const [uncommittedBody, setUncommittedBody] = React.useState(initialBody || '')

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
