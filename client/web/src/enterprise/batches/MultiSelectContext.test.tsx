import { fail } from 'assert'

import React, { useContext, useEffect } from 'react'

import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'
import { act } from 'react-dom/test-utils'

import { MultiSelectContext, MultiSelectContextProvider, type MultiSelectContextState } from './MultiSelectContext'

describe('MultiSelectContextProvider', () => {
    test('providers are initially empty', () => {
        const getContext = mountContext()

        const { selected } = getContext()
        if (selected === 'all') {
            fail()
        }

        expect(selected.size).toBe(0)
    })

    test('selecting and deselecting single IDs works', () => {
        const getContext = mountContext()

        mutateAndAssert(
            getContext,
            context => {
                context.setVisible(true, ['1', '2'])
                context.selectSingle('1')
            },
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(1)
                expect(isSelected('1')).toBe(true)
                expect(isSelected('2')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        mutateAndAssert(
            getContext,
            context => context.selectSingle('2'),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(2)
                expect(isSelected('1')).toBe(true)
                expect(isSelected('2')).toBe(true)
                expect(areAllVisibleSelected()).toBe(true)
            }
        )

        mutateAndAssert(
            getContext,
            context => context.deselectSingle('1'),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(1)
                expect(isSelected('1')).toBe(false)
                expect(isSelected('2')).toBe(true)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )
    })

    test('selecting and deselecting visible works', () => {
        const getContext = mountContext()

        mutateAndAssert(
            getContext,
            context => context.setVisible(true, ['1', '2']),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(0)
                expect(isSelected('1')).toBe(false)
                expect(isSelected('2')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        // Repeat the test twice to ensure it's idempotent.
        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.selectVisible(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(2)
                    expect(isSelected('1')).toBe(true)
                    expect(isSelected('2')).toBe(true)
                    expect(areAllVisibleSelected()).toBe(true)
                }
            )
        )

        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.deselectVisible(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(0)
                    expect(isSelected('1')).toBe(false)
                    expect(isSelected('2')).toBe(false)
                    expect(areAllVisibleSelected()).toBe(false)
                }
            )
        )
    })

    test('selecting and deselecting additional visible works', () => {
        const getContext = mountContext()

        // initial visible data
        mutateAndAssert(
            getContext,
            context => context.setVisible(false, ['1', '2']),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(0)
                expect(isSelected('1')).toBe(false)
                expect(isSelected('2')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        // additional visible data loaded after the fact
        mutateAndAssert(
            getContext,
            context => context.setVisible(false, ['3', '4']),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(0)
                expect(isSelected('3')).toBe(false)
                expect(isSelected('4')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        // Repeat the test twice to ensure it's idempotent.
        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.selectVisible(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(4)
                    expect(isSelected('1')).toBe(true)
                    expect(isSelected('2')).toBe(true)
                    expect(isSelected('3')).toBe(true)
                    expect(isSelected('4')).toBe(true)
                    expect(areAllVisibleSelected()).toBe(true)
                }
            )
        )

        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.deselectVisible(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(0)
                    expect(isSelected('1')).toBe(false)
                    expect(isSelected('2')).toBe(false)
                    expect(isSelected('3')).toBe(false)
                    expect(isSelected('4')).toBe(false)
                    expect(areAllVisibleSelected()).toBe(false)
                }
            )
        )
    })

    test('selecting and deselecting additional visible after filter works', () => {
        const getContext = mountContext()

        // initial loaded data
        mutateAndAssert(
            getContext,
            context => context.setVisible(false, ['1', '2']),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(0)
                expect(isSelected('1')).toBe(false)
                expect(isSelected('2')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        // additional data loaded after a filter
        mutateAndAssert(
            getContext,
            context => context.setVisible(true, ['3', '4']),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(0)
                expect(isSelected('3')).toBe(false)
                expect(isSelected('4')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        // Repeat the test twice to ensure it's idempotent.
        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.selectVisible(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(2)
                    expect(isSelected('3')).toBe(true)
                    expect(isSelected('4')).toBe(true)
                    expect(areAllVisibleSelected()).toBe(true)
                }
            )
        )

        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.deselectVisible(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(0)
                    expect(isSelected('3')).toBe(false)
                    expect(isSelected('4')).toBe(false)
                    expect(areAllVisibleSelected()).toBe(false)
                }
            )
        )
    })

    test('selecting and deselecting all works', () => {
        const getContext = mountContext()

        mutateAndAssert(
            getContext,
            context => context.setVisible(true, ['1', '2']),
            ({ selected }) => {
                expect(selected).not.toBe('all')
            }
        )

        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.selectAll(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    expect(selected).toBe('all')
                    expect(isSelected('1')).toBe(true)
                    expect(isSelected('2')).toBe(true)
                    expect(areAllVisibleSelected()).toBe(true)
                }
            )
        )

        repeat(2, () =>
            mutateAndAssert(
                getContext,
                context => context.deselectAll(),
                ({ areAllVisibleSelected, isSelected, selected }) => {
                    if (selected === 'all') {
                        fail()
                    }

                    expect(selected.size).toBe(0)
                    expect(isSelected('1')).toBe(false)
                    expect(isSelected('2')).toBe(false)
                    expect(areAllVisibleSelected()).toBe(false)
                }
            )
        )

        // Now let's reselect all, let that settle, and test some of the funkier
        // combinations around selecting and deselecting other items or
        // collections.
        act(() => getContext().selectAll())
        mutateAndAssert(
            getContext,
            context => context.deselectVisible(),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(0)
                expect(isSelected('1')).toBe(false)
                expect(isSelected('2')).toBe(false)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        act(() => getContext().selectAll())
        mutateAndAssert(
            getContext,
            context => context.selectVisible(),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(2)
                expect(isSelected('1')).toBe(true)
                expect(isSelected('2')).toBe(true)
                expect(areAllVisibleSelected()).toBe(true)
            }
        )

        act(() => getContext().selectAll())
        mutateAndAssert(
            getContext,
            context => context.deselectSingle('1'),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                if (selected === 'all') {
                    fail()
                }

                expect(selected.size).toBe(1)
                expect(isSelected('1')).toBe(false)
                expect(isSelected('2')).toBe(true)
                expect(areAllVisibleSelected()).toBe(false)
            }
        )

        act(() => getContext().selectAll())
        mutateAndAssert(
            getContext,
            context => context.selectSingle('2'),
            ({ areAllVisibleSelected, isSelected, selected }) => {
                expect(selected).toBe('all')
                expect(isSelected('1')).toBe(true)
                expect(isSelected('2')).toBe(true)
                expect(areAllVisibleSelected()).toBe(true)
            }
        )
    })
})

const mountContext = (): (() => MultiSelectContextState) => {
    let context: MultiSelectContextState | undefined
    render(
        <MultiSelectContextProvider>
            <Reflektor onContext={inner => (context = inner)} />
        </MultiSelectContextProvider>
    )

    return () => {
        if (context === undefined) {
            throw new Error('context did not get populated')
        }
        return context
    }
}

const mutateAndAssert = (
    getContext: () => MultiSelectContextState,
    mutate: (context: MultiSelectContextState) => void,
    assert: (context: MultiSelectContextState) => void
) => {
    act(() => mutate(getContext()))
    assert(getContext())
}

const repeat = (times: number, test: () => void) => {
    for (let index = 0; index < times; index++) {
        test()
    }
}

const Reflektor: React.FunctionComponent<
    React.PropsWithChildren<{ onContext: (inner: MultiSelectContextState) => void }>
> = ({ onContext }) => {
    const context = useContext(MultiSelectContext)
    useEffect(() => {
        onContext(context)
    }, [context, onContext])

    return <>reflektor</>
}
