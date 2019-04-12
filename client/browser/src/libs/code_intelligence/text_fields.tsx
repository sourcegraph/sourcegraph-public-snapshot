import React from 'react'
import { render } from 'react-dom'
import { animationFrameScheduler, fromEvent, Observable, Subscription, Unsubscribable } from 'rxjs'
import { observeOn } from 'rxjs/operators'
import { COMMENT_URI_SCHEME } from '../../../../../shared/src/api/client/types/textDocument'
import { EditorCompletionWidget } from '../../../../../shared/src/components/completion/EditorCompletionWidget'
import { EditorTextFieldUtils } from '../../../../../shared/src/components/editorTextField/EditorTextField'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { MutationRecordLike } from '../../shared/util/dom'
import { CodeHost } from './code_intelligence'
import { trackViews } from './views'

/**
 * Defines a text field that is present on a page and exposes operations for manipulating it.
 */
export interface TextField {
    /** The text field HTML element. */
    element: HTMLTextAreaElement
}

/**
 * Handles added and removed text fields according to the {@link CodeHost} configuration.
 */
export function handleTextFields(
    mutations: Observable<MutationRecordLike[]>,
    { extensionsController }: ExtensionsControllerProps,
    { textFieldResolvers }: Pick<CodeHost, 'textFieldResolvers'>
): Unsubscribable {
    /** A stream of added or removed text fields. */
    const textFields = mutations.pipe(
        trackViews(textFieldResolvers || []),
        observeOn(animationFrameScheduler)
    )
    interface TextFieldState {
        subscriptions: Subscription
    }
    /** Map from text field element to the state associated with it (to be updated or removed) */
    const textFieldStates = new Map<HTMLElement, TextFieldState>()

    // Don't use lodash.uniqueId because that makes it harder to hard-code expected URI values in
    // test code (because the URIs would change depending on test execution order).
    let seq = 0
    const nextModelUri = () => `${COMMENT_URI_SCHEME}://${seq++}`

    return textFields.subscribe(textFieldEvent => {
        console.log(`Text field ${textFieldEvent.type}`, { textFieldEvent })

        // Handle added or removed text fields.
        if (textFieldEvent.type === 'added' && !textFieldStates.has(textFieldEvent.element)) {
            const textFieldState: TextFieldState = {
                subscriptions: new Subscription(),
            }
            textFieldStates.set(textFieldEvent.element, textFieldState)

            // Start 2-way syncing the text field with an editor and model.
            textFieldState.subscriptions.add(
                synchronizeTextField({ extensionsController }, nextModelUri, textFieldEvent)
            )

            textFieldEvent.element.classList.add('sg-mounted')
        } else if (textFieldEvent.type === 'removed') {
            const textFieldState = textFieldStates.get(textFieldEvent.element)
            if (textFieldState) {
                textFieldState.subscriptions.unsubscribe()
                textFieldStates.delete(textFieldEvent.element)
            }
        }
    })
}

/**
 * Start 2-way syncing a text field with an editor and model.
 */
function synchronizeTextField(
    { extensionsController }: ExtensionsControllerProps,
    nextModelUri: () => string,
    { element }: TextField
): Unsubscribable {
    const {
        services: { editor: editorService, model: modelService },
    } = extensionsController

    const subscriptions = new Subscription()

    // Create the editor backing this text field.
    const modelUri = nextModelUri()
    const { text, selections } = EditorTextFieldUtils.getEditorDataFromElement(element)
    modelService.addModel({ uri: modelUri, languageId: 'plaintext', text })
    const editor = editorService.addEditor({
        type: 'CodeEditor',
        resource: modelUri,
        selections,
        isActive: true,
    })
    subscriptions.add(() => editorService.removeEditor(editor))

    // Keep the text field in sync with the editor and model.
    subscriptions.add(
        fromEvent(element, 'input').subscribe(() => {
            EditorTextFieldUtils.updateModelFromElement(modelService, modelUri, element)
            EditorTextFieldUtils.updateEditorSelectionFromElement(editorService, editor, element)
        })
    )
    subscriptions.add(
        fromEvent(element, 'keydown').subscribe(() => {
            EditorTextFieldUtils.updateEditorSelectionFromElement(editorService, editor, element)
        })
    )
    subscriptions.add(
        EditorTextFieldUtils.updateElementOnEditorOrModelChanges(
            editorService,
            editor,
            text => {
                element.value = text
            },
            { current: element }
        )
    )

    // Show completions in the text field.
    const completionWidgetMount = document.createElement('div')
    completionWidgetMount.classList.add('sg-text-field-editor-completion-widget')
    element.insertAdjacentElement('beforebegin', completionWidgetMount)
    render(
        <EditorCompletionWidget
            textArea={element}
            editorId={editor.editorId}
            extensionsController={extensionsController}
        />,
        completionWidgetMount
    )
    subscriptions.add(() => completionWidgetMount.remove())

    return subscriptions
}
