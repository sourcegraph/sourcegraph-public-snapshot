import { afterEach, beforeAll, beforeEach, describe, expect, it } from '@jest/globals'
import { render, type RenderResult, cleanup, fireEvent } from '@testing-library/react'
import sinon from 'sinon'

import { PageSelector, type PageSelectorProps } from './PageSelector'

describe('PageSelector', () => {
    let queries: RenderResult
    const renderWithProps = (props: PageSelectorProps): RenderResult => render(<PageSelector {...props} />)
    const onPageChangeMock = sinon.spy()

    beforeEach(() => {
        onPageChangeMock.resetHistory()
    })

    afterEach(cleanup)

    describe('Invalid configuration', () => {
        it('will error when less than 1 max pages', () => {
            expect(() => {
                renderWithProps({ currentPage: 0, totalPages: 0, onPageChange: onPageChangeMock })
            }).toThrowError(
                ['totalPages must have a value greater than 0', 'currentPage must have a value greater than 0'].join(
                    '\n'
                )
            )
        })

        it('will error when currentPage is less than 1', () => {
            expect(() => {
                renderWithProps({ currentPage: -1, totalPages: 10, onPageChange: onPageChangeMock })
            }).toThrowError('currentPage must have a value greater than 0')
        })

        it('will error when currentPage is greater than totalPages', () => {
            expect(() => {
                renderWithProps({ currentPage: 11, totalPages: 10, onPageChange: onPageChangeMock })
            }).toThrowError('currentPage must be not be greater than totalPages')
        })
    })

    describe('Typical pagination', () => {
        let currentPage = 2
        beforeEach(() => {
            queries = renderWithProps({ currentPage, totalPages: 10, onPageChange: onPageChangeMock })
        })

        it('will render correct elipsis', () => {
            expect(queries.getByText('...', { selector: 'button' })).toBeInTheDocument()
        })

        it('will render correct pages', () => {
            const expectedPages = ['1', '2', '3', '4', '5', '10']
            for (const page of expectedPages) {
                expect(queries.getByRole('button', { name: `Go to page ${page}` })).toBeInTheDocument()
            }
        })

        it('will render correct currently selected page', () => {
            expect(queries.getByRole('button', { name: `Go to page ${currentPage}` })).toHaveAttribute(
                'aria-current',
                'true'
            )
        })

        it('will render previous and next buttons', () => {
            expect(queries.getByText('Previous')).toBeInTheDocument()
            expect(queries.getByText('Next')).toBeInTheDocument()
        })

        it('calls onPageChange correctly when page is individually selected', () => {
            fireEvent.click(queries.getByRole('button', { name: 'Go to page 5' }))
            sinon.assert.calledOnce(onPageChangeMock)
            sinon.assert.calledWith(onPageChangeMock, 5)
        })

        it('calls onPageChange correctly when previous button is selected', () => {
            fireEvent.click(queries.getByText('Previous'))
            sinon.assert.calledOnce(onPageChangeMock)
            sinon.assert.calledWith(onPageChangeMock, currentPage - 1)
        })

        it('calls onPageChange correctly when next button is selected', () => {
            fireEvent.click(queries.getByText('Next'))
            sinon.assert.calledOnce(onPageChangeMock)
            sinon.assert.calledWith(onPageChangeMock, currentPage + 1)
        })

        describe('when currently selected near middle', () => {
            beforeAll(() => {
                currentPage = 5
            })

            it('will render correct elipsis', () => {
                expect(queries.getAllByText('...', { selector: 'button' })).toHaveLength(2)
            })

            it('will render correct pages', () => {
                const expectedPages = ['1', '4', '5', '6', '10']
                for (const page of expectedPages) {
                    expect(queries.getByRole('button', { name: `Go to page ${page}` })).toBeInTheDocument()
                }
            })
        })

        describe('when currently selected near end', () => {
            beforeAll(() => {
                currentPage = 10
            })

            it('will render correct elipsis', () => {
                expect(queries.getByText('...', { selector: 'button' })).toBeInTheDocument()
            })

            it('will render correct pages', () => {
                const expectedPages = ['1', '6', '7', '8', '9', '10']
                for (const page of expectedPages) {
                    expect(queries.getByRole('button', { name: `Go to page ${page}` })).toBeInTheDocument()
                }
            })
        })
    })

    describe('Small pagination', () => {
        const currentPage = 2
        beforeEach(() => {
            queries = renderWithProps({ currentPage, totalPages: 3, onPageChange: onPageChangeMock })
        })

        it('will render no elipsis', () => {
            expect(queries.queryByText('...', { selector: 'button' })).not.toBeInTheDocument()
        })

        it('will render correct pages', () => {
            const expectedPages = ['1', '2', '3']
            for (const page of expectedPages) {
                expect(queries.getByRole('button', { name: `Go to page ${page}` })).toBeInTheDocument()
            }
        })

        it('will render correct currently selected page', () => {
            expect(queries.getByRole('button', { name: `Go to page ${currentPage}` })).toHaveAttribute(
                'aria-current',
                'true'
            )
        })
    })

    describe('Large pagination', () => {
        const currentPage = 1
        beforeEach(() => {
            queries = renderWithProps({ currentPage, totalPages: 30, onPageChange: onPageChangeMock })
        })

        it('will render correct elipsis', () => {
            expect(queries.getByText('...', { selector: 'button' })).toBeInTheDocument()
        })

        it('will render correct pages', () => {
            const expectedPages = ['1', '2', '3', '4', '5', '30']
            for (const page of expectedPages) {
                expect(queries.getByRole('button', { name: `Go to page ${page}` })).toBeInTheDocument()
            }
        })

        it('will render correct currently selected page', () => {
            expect(queries.getByRole('button', { name: `Go to page ${currentPage}` })).toHaveAttribute(
                'aria-current',
                'true'
            )
        })
    })
})
