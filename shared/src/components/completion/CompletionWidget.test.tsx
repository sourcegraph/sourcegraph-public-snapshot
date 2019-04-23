import React from 'react'
// tslint:disable-next-line: no-submodule-imports
import { createRenderer } from 'react-test-renderer/shallow'
import { CompletionItem } from 'sourcegraph'
import { CompletionWidget } from './CompletionWidget'

// tslint:disable: jsx-no-lambda

const COMPLETION_ITEM_2: CompletionItem = { label: 'b' }

describe('CompletionWidgetDropdown', () => {
    test('simple', () => {
        const textArea = document.createElement('textarea')

        const renderer = createRenderer()
        renderer.render(
            <CompletionWidget
                completionListOrError={{ items: [{ label: 'a' }, COMPLETION_ITEM_2] }}
                onSelectItem={() => void 0}
                textArea={textArea}
                listClassName="list-class-name"
                listItemClassName="list-item-class-name"
                loadingClassName="loading-class-name"
                noResultsClassName="no-results-class-name"
                selectedListItemClassName="selected-list-item-class-name"
                widgetClassName="widget-class-name"
                widgetContainerClassName="widget-container-class-name"
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
})
