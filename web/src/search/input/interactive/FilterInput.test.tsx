import React from 'react'
import { FilterInput } from './FilterInput'
import { FilterType, FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import sinon from 'sinon'
import { render, fireEvent, cleanup, getByText } from '@testing-library/react'

const defaultFiltersInQuery: FiltersToTypeAndValue = {
    fork: {
        type: FilterType.fork,
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
    filterType: FilterType.repo as Exclude<FilterType, FilterType.patterntype>,
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
    let nextFiltersInQuery: FiltersToTypeAndValue
    let nextValue: string
    beforeEach(() => {
        nextFiltersInQuery = {}
        nextValue = ''
    })
    const filterHandler = (newFiltersInQuery: FiltersToTypeAndValue, value: string) => {
        nextFiltersInQuery = newFiltersInQuery
        nextValue = value
    }

    const onFilterEditedHandler = (filterKey: string, inputValue: string) => {
        const newFiltersInQuery = {
            ...defaultFiltersInQuery,
            [filterKey]: {
                ...defaultFiltersInQuery[filterKey],
                inputValue,
                editable: false,
            },
        }
        filterHandler(newFiltersInQuery, `${inputValue}`)
    }

    test('Updating filters with an empty value does not work', () => {
        container = render(<FilterInput {...defaultProps} />).container
        const inputEl = container.querySelector('.filter-input__input-field')
        expect(inputEl).toBeTruthy()
        const confirmBtn = container.querySelector('.check-button__btn')
        expect(confirmBtn).toBeTruthy()
        fireEvent.click(confirmBtn!)
        expect(defaultProps.onFilterEdited.notCalled).toBe(true)
    })

    describe('For type filters', () => {
        it('successfully updates when submitting with an empty value', () => {
            container = render(
                <FilterInput {...defaultProps} value="diff" filterType={FilterType.type} mapKey="type" />
            ).container
            const codeRadioButton = container.querySelector('.e2e-filter-input-radio-button-')
            expect(codeRadioButton).toBeTruthy()
            fireEvent.click(codeRadioButton!)
            const confirmBtn = container.querySelector('.e2e-confirm-filter-button')
            expect(confirmBtn).toBeTruthy()
            fireEvent.click(confirmBtn!)
            expect(defaultProps.onFilterEdited.calledOnce).toBe(true)
        })
    })

    describe('For content and message filters', () => {
        it('gets auto-quoted when not editable', () => {
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="content"
                    filterType={FilterType.content}
                    onFilterEdited={onFilterEditedHandler}
                    editable={true}
                />
            ).container

            const inputEl = container.querySelector('.filter-input__input-field')
            expect(inputEl).toBeTruthy()
            fireEvent.click(inputEl!)
            fireEvent.change(inputEl!, { target: { value: 'test query' } })
            const confirmBtn = container.querySelector('.check-button__btn')
            expect(confirmBtn).toBeTruthy()
            fireEvent.click(confirmBtn!)
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="content"
                    filtersInQuery={nextFiltersInQuery}
                    filterType={FilterType.content}
                    value={nextValue}
                    onFilterEdited={onFilterEditedHandler}
                    editable={false}
                />
            ).container
            expect(getByText(container, 'content:"test query"')).toBeTruthy()
        })
        it('gets stripped of quotes when editable', () => {
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="content"
                    filterType={FilterType.content}
                    onFilterEdited={onFilterEditedHandler}
                    editable={true}
                />
            ).container

            const inputEl = container.querySelector('.filter-input__input-field')
            expect(inputEl).toBeTruthy()
            fireEvent.click(inputEl!)
            fireEvent.change(inputEl!, { target: { value: 'test query' } })
            const confirmBtn = container.querySelector('.check-button__btn')
            expect(confirmBtn).toBeTruthy()
            fireEvent.click(confirmBtn!)
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="content"
                    filtersInQuery={nextFiltersInQuery}
                    filterType={FilterType.content}
                    value={nextValue}
                    onFilterEdited={onFilterEditedHandler}
                    editable={false}
                />
            ).container
            fireEvent.click(container.querySelector('.filter-input__button-text')!)
            expect(getByText(container, 'content:test query')).toBeTruthy()
        })

        test('filter input for message filters does not get auto-quoted when editable', () => {
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="message"
                    filterType={FilterType.message}
                    onFilterEdited={onFilterEditedHandler}
                    editable={true}
                />
            ).container

            const inputEl = container.querySelector('.filter-input__input-field')
            expect(inputEl).toBeTruthy()
            fireEvent.click(inputEl!)
            fireEvent.change(inputEl!, { target: { value: 'test query' } })
            const confirmBtn = container.querySelector('.check-button__btn')
            expect(confirmBtn).toBeTruthy()
            fireEvent.click(confirmBtn!)
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="message"
                    filtersInQuery={nextFiltersInQuery}
                    filterType={FilterType.message}
                    value={nextValue}
                    onFilterEdited={onFilterEditedHandler}
                    editable={false}
                />
            ).container
            expect(getByText(container, 'message:"test query"')).toBeTruthy()
        })

        it('filter input for content filters calls onFilterEdited with quoted value', () => {
            container = render(
                <FilterInput
                    {...defaultProps}
                    mapKey="content"
                    filterType={FilterType.content}
                    onFilterEdited={defaultProps.onFilterEdited}
                    editable={true}
                />
            ).container

            const inputEl = container.querySelector('.filter-input__input-field')
            expect(inputEl).toBeTruthy()
            fireEvent.click(inputEl!)
            fireEvent.change(inputEl!, { target: { value: 'test query' } })
            const confirmBtn = container.querySelector('.check-button__btn')
            expect(confirmBtn).toBeTruthy()
            fireEvent.click(confirmBtn!)
            expect(defaultProps.onFilterEdited.calledWith('content', '"test query"')).toBe(true)
        })
    })
})
