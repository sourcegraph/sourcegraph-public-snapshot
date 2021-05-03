import { render, RenderResult, within, BoundFunction, GetByRole, cleanup, fireEvent } from '@testing-library/react'
import openColor from 'open-color'
import * as React from 'react'
import sinon from 'sinon'

import { asError } from '@sourcegraph/shared/src/util/errors'

import { FORM_ERROR } from '../../hooks/useForm';

import { CreationSearchInsightForm, CreationSearchInsightFormProps } from './CreationSearchInsightForm'

describe('CreateInsightForm', () => {
    const renderWithProps = (props: CreationSearchInsightFormProps): RenderResult =>
        render(<CreationSearchInsightForm {...props} />)
    const onSubmitMock = sinon.spy()

    beforeEach(() => onSubmitMock.resetHistory())
    afterEach(cleanup)

    const getFormFields = (getByRole: BoundFunction<GetByRole>) => {
        const title = getByRole('textbox', { name: /title/i })
        const repositories = getByRole('textbox', { name: /repositories/i })

        const personalVisibility = getByRole('radio', { name: /personal/i })
        const organisationVisibility = getByRole('radio', { name: /organization/i })

        const dataSeriesGroup = getByRole('group', { name: /data series/i })
        const seriesName = within(dataSeriesGroup).getByRole('textbox', { name: /name/i })
        const seriesQuery = within(dataSeriesGroup).getByRole('textbox', { name: /query/i })

        const seriesColorGroup = within(dataSeriesGroup).getByRole('group', { name: /color/i })

        const seriesColorRadioButtons = within(seriesColorGroup).getAllByRole('radio') as HTMLInputElement[]

        const stepGroup = getByRole('group', { name: /step/i })

        const stepValue = within(stepGroup).getByRole('textbox')

        const stepRadioButtons = within(stepGroup).getAllByRole('radio')

        return {
            title,
            repositories,
            personalVisibility,
            organisationVisibility,
            dataSeriesGroup,
            seriesName,
            seriesQuery,
            seriesColorRadioButtons,
            stepValue,
            stepRadioButtons,
        }
    }

    describe('with common fill flow', () => {
        it('will render standard package of form fields', () => {
            const { getByRole } = renderWithProps({ onSubmit: onSubmitMock })

            const {
                title,
                repositories,
                personalVisibility,
                organisationVisibility,
                seriesName,
                seriesQuery,
                seriesColorRadioButtons,
                stepValue,
                stepRadioButtons,
            } = getFormFields(getByRole)

            expect(title).toBeInTheDocument()
            expect(repositories).toBeInTheDocument()
            expect(personalVisibility).toBeInTheDocument()
            expect(organisationVisibility).toBeInTheDocument()
            expect(seriesName).toBeInTheDocument()
            expect(seriesQuery).toBeInTheDocument()

            // Since we use 12 standard open color as a default colors for color picker
            expect(seriesColorRadioButtons.length).toBe(12)

            expect(stepValue).toBeInTheDocument()
            // Since we have 5 options for insight step (hours, days, weeks, months, years)
            expect(stepRadioButtons.length).toBe(5)
        })

        it('will fire onSubmit if all fields have been filled with valid value', () => {
            const { getByRole } = renderWithProps({ onSubmit: onSubmitMock })
            const {
                title,
                repositories,
                organisationVisibility,
                dataSeriesGroup,
                seriesName,
                seriesQuery,
                stepValue,
            } = getFormFields(getByRole)

            fireEvent.change(title, { target: { value: 'First code insight' } })
            fireEvent.change(repositories, { target: { value: 'github.com/sourcegraph/sourcegraph' } })
            fireEvent.click(organisationVisibility)

            const submitSeriesButton = within(dataSeriesGroup).getByRole('button', { name: /done/i })

            const yellowColorRadio = within(dataSeriesGroup).getByRole('radio', { name: /yellow/i })

            fireEvent.change(seriesName, { target: { value: 'First code insight series' } })
            fireEvent.change(seriesQuery, { target: { value: 'patternType:regex case:yes \\*\\sas\\sGQL' } })
            fireEvent.click(yellowColorRadio)
            fireEvent.click(submitSeriesButton)

            const monthsRadio = getByRole('radio', { name: /months/i })

            fireEvent.change(stepValue, { target: { value: 2 } })
            fireEvent.click(monthsRadio)

            const submitButton = getByRole('button', { name: /create code insight/i })

            fireEvent.click(submitButton)

            sinon.assert.calledOnce(onSubmitMock)
            sinon.assert.calledWith(onSubmitMock, {
                title: 'First code insight',
                repositories: 'github.com/sourcegraph/sourcegraph',
                visibility: 'organization',
                series: [
                    {
                        name: 'First code insight series',
                        query: 'patternType:regex case:yes \\*\\sas\\sGQL',
                        // Open color value from our own css variables
                        color: openColor.yellow[7],
                    },
                ],
                stepValue: '2',
                step: 'months',
            })
        })
    })

    describe('show error massage', () => {
        it('with invalid title field', () => {
            const { getByRole, getByText } = renderWithProps({ onSubmit: onSubmitMock })
            const title = getByRole('textbox', { name: /title/i })
            const submitButton = getByRole('button', { name: /create code insight/i })

            fireEvent.click(submitButton)

            sinon.assert.notCalled(onSubmitMock)

            expect(title).toHaveFocus()
            expect(getByText(/title is required/i)).toBeInTheDocument()
        })

        it('with invalid repository field', () => {
            const { getByRole, getByText } = renderWithProps({ onSubmit: onSubmitMock })
            const title = getByRole('textbox', { name: /title/i })
            const repositories = getByRole('textbox', { name: /repositories/i })
            const submitButton = getByRole('button', { name: /create code insight/i })

            fireEvent.change(title, { target: { value: 'First code insight' } })
            fireEvent.click(submitButton)

            sinon.assert.notCalled(onSubmitMock)

            expect(repositories).toHaveFocus()
            expect(getByText(/repositories is required/i)).toBeInTheDocument()
        })

        it('with invalid data series field', () => {
            const { getByRole, getByText } = renderWithProps({ onSubmit: onSubmitMock })
            const title = getByRole('textbox', { name: /title/i })
            const repositories = getByRole('textbox', { name: /repositories/i })
            const submitButton = getByRole('button', { name: /create code insight/i })
            const dataSeriesGroup = getByRole('group', { name: /data series/i })
            const seriesName = within(dataSeriesGroup).getByRole('textbox', { name: /name/i })

            fireEvent.change(title, { target: { value: 'First code insight' } })
            fireEvent.change(repositories, { target: { value: 'github.com/sourcegraph/sourcegraph' } })
            fireEvent.click(submitButton)

            sinon.assert.notCalled(onSubmitMock)

            expect(seriesName).toHaveFocus()
            expect(getByText(/series is empty/i)).toBeInTheDocument()
        })

        it('when onSubmit threw submit error', () => {
            const onSubmit = () => ({ [FORM_ERROR]: asError(new Error('Submit error')) })
            const { getByRole, getByText } = renderWithProps({ onSubmit })
            const {
                title,
                repositories,
                organisationVisibility,
                dataSeriesGroup,
                seriesName,
                seriesQuery,
                stepValue,
            } = getFormFields(getByRole)

            fireEvent.change(title, { target: { value: 'First code insight' } })
            fireEvent.change(repositories, { target: { value: 'github.com/sourcegraph/sourcegraph' } })
            fireEvent.click(organisationVisibility)

            const submitSeriesButton = within(dataSeriesGroup).getByRole('button', { name: /done/i })

            const yellowColorRadio = within(dataSeriesGroup).getByRole('radio', { name: /yellow/i })

            fireEvent.change(seriesName, { target: { value: 'First code insight series' } })
            fireEvent.change(seriesQuery, { target: { value: 'patternType:regex case:yes \\*\\sas\\sGQL' } })
            fireEvent.click(yellowColorRadio)
            fireEvent.click(submitSeriesButton)

            const monthsRadio = getByRole('radio', { name: /months/i })

            fireEvent.change(stepValue, { target: { value: 2 } })
            fireEvent.click(monthsRadio)

            const submitButton = getByRole('button', { name: /create code insight/i })

            fireEvent.click(submitButton)

            expect(getByText(/submit error/i)).toBeInTheDocument()
        })
    })
})
