import { isEqual } from 'lodash'
import React from 'react'
import { Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, map } from 'rxjs/operators'
import { CodeEditorData, ViewerId, ViewerService, observeEditorAndModel } from '../../api/client/services/viewerService'
import { ModelService, TextModel } from '../../api/client/services/modelService'
import { offsetToPosition, positionToOffset } from '../../api/client/types/textDocument'
import { isDefined } from '../../util/types'

/**
 * Utilities for 2-way syncing an HTMLTextAreaElement with an editor and model. These are factored
 * out of the {@link EditorTextField} component so that they can be used in other places, such as in
 * the browser extension.
 */
export const EditorTextFieldUtils = {
    /**
     * Reads the current selections and value from the element.
     */
    getEditorDataFromElement: (
        element: HTMLTextAreaElement
    ): Pick<TextModel, 'text'> & Pick<CodeEditorData, 'selections'> => {
        const isReversed = element.selectionDirection === 'backward'
        const selectionStart = isReversed ? element.selectionEnd : element.selectionStart
        const selectionEnd = isReversed ? element.selectionStart : element.selectionEnd
        const start = offsetToPosition(element.value, selectionStart)
        const end = selectionStart === selectionEnd ? start : offsetToPosition(element.value, selectionEnd)
        return {
            text: element.value,
            selections: [{ anchor: start, active: end, start, end, isReversed }],
        }
    },

    /**
     * Update the editor's selection from the element's selection.
     */
    updateEditorSelectionFromElement: (
        viewerService: Pick<ViewerService, 'setSelections'>,
        editor: ViewerId,
        element: HTMLTextAreaElement
    ): void => {
        viewerService.setSelections(editor, EditorTextFieldUtils.getEditorDataFromElement(element).selections)
    },

    /**
     * Update the model from the element's value.
     */
    updateModelFromElement: (
        modelService: Pick<ModelService, 'updateModel'>,
        modelUri: string,
        element: HTMLTextAreaElement
    ): void => {
        modelService.updateModel(modelUri, element.value)
    },

    /**
     * Update the element's value (via {@link setValue}) and selection range whenever the editor or
     * model change.
     */
    updateElementOnEditorOrModelChanges: (
        viewerService: Pick<ViewerService, 'observeViewer'>,
        modelService: Pick<ModelService, 'observeModel'>,
        editor: ViewerId,
        setValue: (text: string) => void,
        textAreaReference: React.RefObject<Pick<HTMLTextAreaElement, 'value' | 'setSelectionRange'>>
    ): Unsubscribable => {
        const subscriptions = new Subscription()

        const changes = observeEditorAndModel(editor, viewerService, modelService)
        const modelTextChanges = changes.pipe(
            map(({ model: { text } }) => text),
            filter(isDefined),
            distinctUntilChanged()
        )

        // Update text.
        subscriptions.add(modelTextChanges.subscribe(text => setValue(text)))

        // Update selection.
        subscriptions.add(
            changes
                .pipe(
                    map(editor => editor.selections),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    filter(selections => selections.length !== 0)
                )
                .subscribe(selections => {
                    const textArea = textAreaReference.current
                    if (textArea) {
                        const sel = selections[0] // TODO: Only a single selection is supported.
                        const start = positionToOffset(textArea.value, sel.start)
                        const isEmpty = sel.start.line === sel.end.line && sel.start.character === sel.end.character
                        const end = isEmpty ? start : positionToOffset(textArea.value, sel.end)
                        textArea.setSelectionRange(
                            sel.isReversed ? end : start,
                            sel.isReversed ? start : end,
                            sel.isReversed ? 'backward' : 'forward'
                        )
                    }
                })
        )

        return subscriptions
    },
}
