import H from 'history'
import { uniqueId } from 'lodash'
import React, { createRef, useCallback, useEffect, useLayoutEffect, useMemo, useState } from 'react'
import { Key } from 'ts-key-enum'
import { TextModel } from '../../../../shared/src/api/client/services/modelService'
import { COMMENT_URI_SCHEME } from '../../../../shared/src/api/client/types/textDocument'
import { EditorTextField } from '../../../../shared/src/components/editorTextField/EditorTextField'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { Form } from '../../components/Form'
import { WebEditorCompletionWidget } from '../../components/shared'

interface Props extends ExtensionsControllerProps {
    /** The initial body (used when editing an existing comment). */
    initialBody?: string

    placeholder: string

    /** The label to display on the submit button. */
    submitLabel: string

    /** Called when the submit button is clicked. */
    onSubmit: (body: string) => void

    /**
     * If set, a "Cancel" button is shown, and this callback is called when it is clicked.
     */
    onCancel?: () => void

    autoFocus?: boolean
    disabled?: boolean
    className?: string
    // TODO!(sqs): confirm navigation away when field is dirty
    history: H.History
}

/**
 * A form to create or edit a comment.
 */
export const CommentForm: React.FunctionComponent<Props> = ({
    initialBody,
    submitLabel,
    placeholder,
    onSubmit,
    onCancel,
    autoFocus,
    disabled,
    className = '',
    history,
    extensionsController,
}) => {
    const [uncommittedBody, setUncommittedBody] = useState(initialBody || '')

    const onFormSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            onSubmit(uncommittedBody)
        },
        [onSubmit, uncommittedBody]
    )

    // Warn when navigating away from page when that would result in loss of user input.
    useEffect(() => {
        const isDirty = uncommittedBody !== (initialBody || '')
        if (isDirty) {
            return history.block('Discard unsaved comment?')
        }
        return undefined
    }, [history, initialBody, uncommittedBody])

    // Text field completion.
    const [textArea, setTextArea] = useState<HTMLTextAreaElement | null>(null)
    const textAreaRef = createRef<HTMLTextAreaElement>()
    useLayoutEffect(() => setTextArea(textAreaRef.current), [textAreaRef])
    const { editorId, modelUri } = useMemo(() => {
        const model: TextModel = {
            uri: uniqueId(`${COMMENT_URI_SCHEME}://`),
            languageId: 'plaintext',
            text: initialBody || '',
        }
        extensionsController.services.model.addModel(model)
        const editor = extensionsController.services.editor.addEditor({
            type: 'CodeEditor',
            resource: model.uri,
            selections: [],
            isActive: true,
        })
        return { editorId: editor.editorId, modelUri: model.uri }
    }, [extensionsController.services.editor, extensionsController.services.model, initialBody])
    useEffect(
        () => () => {
            extensionsController.services.editor.removeEditor({ editorId })
            extensionsController.services.model.removeModel(modelUri)
        },
        [editorId, extensionsController.services.editor, extensionsController.services.model, modelUri]
    )

    // Ctrl/Meta+Enter to submit.
    const onKeyDown = useCallback<React.KeyboardEventHandler<HTMLTextAreaElement>>(
        e => {
            if ((e.ctrlKey || e.metaKey) && e.key === Key.Enter) {
                onSubmit(uncommittedBody)
            }
        },
        [onSubmit, uncommittedBody]
    )

    return (
        <Form className={`comment-form ${className}`} onSubmit={onFormSubmit}>
            {textArea && (
                <WebEditorCompletionWidget
                    textArea={textArea}
                    editorId={editorId}
                    extensionsController={extensionsController}
                />
            )}
            <EditorTextField
                className="form-control mb-2"
                placeholder={placeholder}
                editorId={editorId}
                modelUri={modelUri}
                onValueChange={setUncommittedBody}
                onKeyDown={onKeyDown}
                textAreaRef={textAreaRef}
                autoFocus={autoFocus}
                rows={5} // TODO!(sqs): use autosizing textarea and make this minRows={5}
                disabled={disabled}
                extensionsController={extensionsController}
                style={{ resize: 'vertical' }}
            />
            <div className="d-flex align-items-center justify-content-end">
                {onCancel && (
                    <button type="reset" className="btn btn-link" disabled={disabled} onClick={onCancel}>
                        Cancel
                    </button>
                )}
                <button type="submit" className="btn btn-primary" disabled={disabled}>
                    {submitLabel}
                </button>
            </div>
        </Form>
    )
}
