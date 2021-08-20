import { mount } from 'enzyme'
import React from 'react'
import { act } from 'react-dom/test-utils'
import { ButtonDropdown } from 'reactstrap'
import sinon from 'sinon'

import { Progress } from '@sourcegraph/shared/src/search/stream'

import { StreamingProgressSkippedButton } from './StreamingProgressSkippedButton'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

describe('StreamingProgressSkippedButton', () => {
    it('should not show if no skipped items', () => {
        const progress: Progress = {
            durationMs: 0,
            matchCount: 0,
            skipped: [],
        }

        const element = mount(<StreamingProgressSkippedButton progress={progress} onSearchAgain={sinon.spy()} />)
        expect(element.find('.streaming-progress__skipped')).toHaveLength(0)
        expect(element.find('.streaming-progress__skipped-popover')).toHaveLength(0)
    })

    it('should be in info state with only info items', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedButton progress={progress} onSearchAgain={sinon.spy()} />)
        expect(element.find('.btn.btn-outline-secondary.streaming-progress__skipped')).toHaveLength(1)
        expect(element.find('.btn.btn-outline-danger.streaming-progress__skipped')).toHaveLength(0)
    })

    it('should be in warning state with at least one warning item', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedButton progress={progress} onSearchAgain={sinon.spy()} />)
        expect(element.find('.btn.btn-outline-danger.streaming-progress__skipped')).toHaveLength(1)
        expect(element.find('.btn.btn-outline-secondary.streaming-progress__skipped')).toHaveLength(0)
    })

    it('should open and close popover when button is clicked', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedButton progress={progress} onSearchAgain={sinon.spy()} />)

        let popover = element.find(ButtonDropdown)
        expect(popover.prop('isOpen')).toBe(false)

        const button = element.find('.btn.streaming-progress__skipped')
        button.simulate('click')

        popover = element.find(ButtonDropdown)
        expect(popover.prop('isOpen')).toBe(true)

        button.simulate('click')

        popover = element.find(ButtonDropdown)
        expect(popover.prop('isOpen')).toBe(false)
    })

    it('should close popup and call onSearchAgain callback when popover raises event', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
            ],
        }

        const onSearchAgain = sinon.spy()

        const element = mount(<StreamingProgressSkippedButton progress={progress} onSearchAgain={onSearchAgain} />)

        // Open dropdown
        const button = element.find('.btn.streaming-progress__skipped')
        button.simulate('click')

        // Trigger onSearchAgain event and check for changes
        const skippedPopover = element.find(StreamingProgressSkippedPopover)
        act(() => skippedPopover.prop('onSearchAgain')(['archived:yes']))
        element.update()

        const dropdown = element.find(ButtonDropdown)
        expect(dropdown.prop('isOpen')).toBe(false)

        sinon.assert.calledOnce(onSearchAgain)
        sinon.assert.calledWith(onSearchAgain, ['archived:yes'])
    })
})
