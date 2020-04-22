import React from 'react'
import { render } from 'react-dom'
import { animationFrameScheduler, fromEvent, Observable, Subscription, Unsubscribable } from 'rxjs'
import { observeOn } from 'rxjs/operators'
import { COMMENT_URI_SCHEME } from '../../../../shared/src/api/client/types/textDocument'
import { EditorCompletionWidget } from '../../../../shared/src/components/completion/EditorCompletionWidget'
import { EditorTextFieldUtils } from '../../../../shared/src/components/editorTextField/EditorTextField'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
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
    {
        textFieldResolvers,
        completionWidgetClassProps,
    }: Pick<CodeHost, 'textFieldResolvers' | 'completionWidgetClassProps'>
): Unsubscribable {
    /** A stream of added or removed text fields. */
    const textFields = mutations.pipe(trackViews(textFieldResolvers || []), observeOn(animationFrameScheduler))

    // Don't use lodash.uniqueId because that makes it harder to hard-code expected URI values in
    // test code (because the URIs would change depending on test execution order).
    let seq = 0
    const nextModelUri = (): string => `${COMMENT_URI_SCHEME}://${seq++}`

    return textFields.subscribe(textFieldEvent => {
        console.log('Text field added', { textFieldEvent })
        textFieldEvent.subscriptions.add(() => console.log('Text field removed', { textFieldEvent }))
        // Start 2-way syncing the text field with an editor and model.
        textFieldEvent.subscriptions.add(
            synchronizeTextField({ extensionsController }, { completionWidgetClassProps }, nextModelUri, textFieldEvent)
        )
    })
}

/**
 * Start 2-way syncing a text field with an editor and model.
 */
function synchronizeTextField(
    { extensionsController }: ExtensionsControllerProps,
    { completionWidgetClassProps }: Pick<CodeHost, 'completionWidgetClassProps'>,
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
        fromEvent(element, 'input')
            .pipe(observeOn(animationFrameScheduler))
            .subscribe(() => {
                EditorTextFieldUtils.updateModelFromElement(modelService, modelUri, element)
                EditorTextFieldUtils.updateEditorSelectionFromElement(editorService, editor, element)
            })
    )
    subscriptions.add(
        fromEvent(element, 'keydown')
            .pipe(observeOn(animationFrameScheduler))
            .subscribe(() => {
                EditorTextFieldUtils.updateEditorSelectionFromElement(editorService, editor, element)
            })
    )
    subscriptions.add(
        EditorTextFieldUtils.updateElementOnEditorOrModelChanges(
            editorService,
            modelService,
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
            {...completionWidgetClassProps}
            textArea={element}
            editorId={editor.editorId}
            extensionsController={extensionsController}
        />,
        completionWidgetMount
    )
    subscriptions.add(() => completionWidgetMount.remove())

    return subscriptions
}
