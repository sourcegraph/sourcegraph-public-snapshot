import React from 'react'
import { render, RenderResult, cleanup, fireEvent } from '@testing-library/react'
import { PageSelector, Props } from './PageSelector'

describe('PageSelector', () => {
    let queries: RenderResult
    const renderWithProps = (props: Props): RenderResult => render(<PageSelector {...props} />)
    const onPageChangeMock = jest.fn()

    beforeEach(() => {
        onPageChangeMock.mockReset()
    })

    afterEach(cleanup)

    describe('Invalid configuration', () => {
        it('will error when less than 1 max pages', () => {
            expect(() => {
                renderWithProps({ currentPage: 0, maxPages: 0, onPageChange: onPageChangeMock })
            }).toThrowErrorMatchingSnapshot()
        })

        it('will error when currentPage is less than 1', () => {
            expect(() => {
                renderWithProps({ currentPage: -1, maxPages: 10, onPageChange: onPageChangeMock })
            }).toThrowErrorMatchingSnapshot()
        })

        it('will error when currentPage is greater than maxPages', () => {
            expect(() => {
                renderWithProps({ currentPage: 11, maxPages: 10, onPageChange: onPageChangeMock })
            }).toThrowErrorMatchingSnapshot()
        })
    })

    describe('Typical pagination', () => {
        let currentPage = 2
        beforeEach(() => {
            queries = renderWithProps({ currentPage, maxPages: 10, onPageChange: onPageChangeMock })
        })

        it('will render correct elipsis', () => {
            expect(queries.getAllByRole('button', { name: '...' })).toHaveLength(1)
        })

        it('will render correct pages', () => {
            const expectedPages = ['1', '2', '3', '4', '5', '10']
            expectedPages.forEach(page => {
                expect(queries.getByRole('button', { name: page })).toBeInTheDocument()
            })
        })

        it('will render correct currently selected page', () => {
            expect(queries.getByRole('button', { name: String(currentPage) })).toHaveAttribute('aria-current', 'true')
        })

        it('will render previous and next buttons', () => {
            expect(queries.getByText('Previous')).toBeInTheDocument()
            expect(queries.getByText('Next')).toBeInTheDocument()
        })

        it('calls onPageChange correctly when page is individually selected', () => {
            fireEvent.click(queries.getByRole('button', { name: '5' }))
            expect(onPageChangeMock).toHaveBeenCalledTimes(1)
            expect(onPageChangeMock).toHaveBeenCalledWith(5)
        })

        it('calls onPageChange correctly when previous button is selected', () => {
            fireEvent.click(queries.getByText('Previous'))
            expect(onPageChangeMock).toHaveBeenCalledTimes(1)
            expect(onPageChangeMock).toHaveBeenCalledWith(currentPage - 1)
        })

        it('calls onPageChange correctly when next button is selected', () => {
            fireEvent.click(queries.getByText('Next'))
            expect(onPageChangeMock).toHaveBeenCalledTimes(1)
            expect(onPageChangeMock).toHaveBeenCalledWith(currentPage + 1)
        })

        describe('rendering near middle', () => {
            beforeAll(() => {
                currentPage = 5
            })

            it('will render correct elipsis', () => {
                expect(queries.getAllByRole('button', { name: '...' })).toHaveLength(2)
            })

            it('will render correct pages', () => {
                const expectedPages = ['1', '4', '5', '6', '10']
                expectedPages.forEach(page => {
                    expect(queries.getByRole('button', { name: page })).toBeInTheDocument()
                })
            })
        })
    })

    describe('Small pagination', () => {
        const currentPage = 2
        beforeEach(() => {
            queries = renderWithProps({ currentPage, maxPages: 3, onPageChange: onPageChangeMock })
        })

        it('will render no elipsis', () => {
            expect(queries.queryAllByRole('button', { name: '...' })).toHaveLength(0)
        })

        it('will render correct pages', () => {
            const expectedPages = ['1', '2', '3']
            expectedPages.forEach(page => {
                expect(queries.getByRole('button', { name: page })).toBeInTheDocument()
            })
        })

        it('will render correct currently selected page', () => {
            expect(queries.getByRole('button', { name: String(currentPage) })).toHaveAttribute('aria-current', 'true')
        })
    })

    describe('Large pagination', () => {
        const currentPage = 1
        beforeEach(() => {
            queries = renderWithProps({ currentPage, maxPages: 30, onPageChange: onPageChangeMock })
        })

        it('will render correct elipsis', () => {
            expect(queries.getAllByRole('button', { name: '...' })).toHaveLength(1)
        })

        it('will render correct pages', () => {
            const expectedPages = ['1', '2', '3', '4', '5', '30']
            expectedPages.forEach(page => {
                expect(queries.getByRole('button', { name: page })).toBeInTheDocument()
            })
        })

        it('will render correct currently selected page', () => {
            expect(queries.getByRole('button', { name: String(currentPage) })).toHaveAttribute('aria-current', 'true')
        })
    })
})
