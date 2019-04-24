import React from 'react'
import renderer from 'react-test-renderer'
import { CompletionItem } from 'sourcegraph'
import { CompletionWidgetDropdown } from './CompletionWidgetDropdown'

// tslint:disable: jsx-no-lambda

const COMPLETION_ITEM_2: CompletionItem = { label: 'b' }

describe('CompletionWidgetDropdown', () => {
    test('simple', () =>
        expect(
            renderer
                .create(
                    <CompletionWidgetDropdown
                        completionListOrError={{ items: [{ label: 'a' }, COMPLETION_ITEM_2] }}
                        highlightedIndex={1}
                        onClickOutside={() => void 0}
                        onDownshiftStateChange={() => void 0}
                        onItemSelected={() => void 0}
                        selectedItem={COMPLETION_ITEM_2}
                        listClassName="list-class-name"
                        listItemClassName="list-item-class-name"
                        loadingClassName="loading-class-name"
                        noResultsClassName="no-results-class-name"
                        selectedListItemClassName="selected-list-item-class-name"
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
