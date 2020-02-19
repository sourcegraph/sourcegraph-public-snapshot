import React from 'react'
import { FilterInput } from './FilterInput'
import { FilterTypes, FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import sinon from 'sinon'
import { render, fireEvent, cleanup, getByText, getByDisplayValue } from '@testing-library/react'

const defaultFiltersInQuery: FiltersToTypeAndValue = {
    fork: {
        type: FilterTypes.fork,
        value: 'no',
        editable: false,
        negated: false,
    },
}
const defaultProps = {
    filtersInQuery: defaultFiltersInQuery,
    navbarQuery: { query: 'test', cursorPosition: 4 },
    mapKey: 'repo',
    value: '',
    filterType: FilterTypes.repo,
    editable: true,
    negated: false,
    isHomepage: false,
    onSubmit: sinon.spy(),
    onFilterEdited: sinon.spy(),
    onFilterDeleted: sinon.spy(),
    toggleFilterEditable: sinon.spy(),
    toggleFilterNegated: sinon.spy(),
}

describe('FilterInput', () => {
    afterAll(cleanup)
    let container: HTMLElement
    beforeEach(() => {
        ;({ container } = render(<FilterInput {...defaultProps} editable={true} />))
    })

    let nextFiltersInQuery = {}
    let nextValue = ''
    const filterHandler = (newFiltersInQuery: FiltersToTypeAndValue, value: string) => {
        nextFiltersInQuery = newFiltersInQuery
        nextValue = value
    }

    const onFilterEditedHandler = async (filterKey: string, inputValue: string) => {
        const newFiltersInQuery = {
            ...defaultFiltersInQuery,
            [filterKey]: {
                ...defaultFiltersInQuery[filterKey],
                inputValue,
                editable: false,
            },
        }
        await filterHandler(newFiltersInQuery, `${inputValue}`)
    }

    it('filter input for content filters get auto-quoted', () => {
        ;({ container } = render(
            <FilterInput
                {...defaultProps}
                mapKey="content"
                filterType={FilterTypes.content}
                onFilterEdited={onFilterEditedHandler}
                editable={true}
            />
        ))

        const inputEl = container.querySelector('.filter-input__input-field')
        expect(inputEl).toBeTruthy()
        fireEvent.click(inputEl!)
        fireEvent.change(inputEl!, { target: { value: 'test query' } })
        const confirmBtn = container.querySelector('.check-button__btn')
        expect(confirmBtn).toBeTruthy()
        fireEvent.click(confirmBtn!)
        ;({ container } = render(
            <FilterInput
                {...defaultProps}
                mapKey="content"
                filtersInQuery={nextFiltersInQuery}
                filterType={FilterTypes.content}
                value={nextValue}
                onFilterEdited={onFilterEditedHandler}
                editable={false}
            />
        ))
        expect(getByText(container, 'content:"test query"')).toBeTruthy()
    })

    test('filter input for message filters get auto-quoted', () => {
        ;({ container } = render(
            <FilterInput
                {...defaultProps}
                mapKey="message"
                filterType={FilterTypes.message}
                onFilterEdited={onFilterEditedHandler}
                editable={true}
            />
        ))

        const inputEl = container.querySelector('.filter-input__input-field')
        expect(inputEl).toBeTruthy()
        fireEvent.click(inputEl!)
        fireEvent.change(inputEl!, { target: { value: 'test query' } })
        const confirmBtn = container.querySelector('.check-button__btn')
        expect(confirmBtn).toBeTruthy()
        fireEvent.click(confirmBtn!)
        ;({ container } = render(
            <FilterInput
                {...defaultProps}
                mapKey="message"
                filtersInQuery={nextFiltersInQuery}
                filterType={FilterTypes.message}
                value={nextValue}
                onFilterEdited={onFilterEditedHandler}
                editable={false}
            />
        ))
        expect(getByText(container, 'message:"test query"')).toBeTruthy()
    })

    test('Updating filters with an empty value does not work', () => {
        ;({ container } = render(<FilterInput {...defaultProps} />))
        const inputEl = container.querySelector('.filter-input__input-field')
        expect(inputEl).toBeTruthy()
        const confirmBtn = container.querySelector('.check-button__btn')
        expect(confirmBtn).toBeTruthy()
        fireEvent.click(confirmBtn!)
        expect(defaultProps.onFilterEdited.notCalled).toBe(true)
    })

    test('Updating type filter with an empty value does work', () => {
        ;({ container } = render(
            <FilterInput {...defaultProps} value="diff" filterType={FilterTypes.type} mapKey="type" />
        ))
        const codeRadioButton = container.querySelector('.e2e-filter-input-radio-button-')
        expect(codeRadioButton).toBeTruthy()
        fireEvent.click(codeRadioButton!)
        const confirmBtn = container.querySelector('.e2e-confirm-filter-button')
        expect(confirmBtn).toBeTruthy()
        fireEvent.click(confirmBtn!)
        expect(defaultProps.onFilterEdited.calledOnce).toBe(true)
    })
})
