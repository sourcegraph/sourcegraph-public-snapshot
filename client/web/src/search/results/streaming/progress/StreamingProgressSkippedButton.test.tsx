import { mount } from 'enzyme'
import React from 'react'
import { ButtonDropdown } from 'reactstrap'
import { Progress } from '../../../stream'
import { StreamingProgressSkippedButton } from './StreamingProgressSkippedButton'

describe('StreamingProgressSkippedButton', () => {
    let div: HTMLDivElement
    beforeEach(() => {
        div = document.createElement('div')
        document.body.append(div)
    })

    it('should not show if no skipped items', () => {
        const progress: Progress = {
            done: false,
            durationMs: 0,
            matchCount: 0,
            skipped: [],
        }

        const element = mount(<StreamingProgressSkippedButton progress={progress} />, { attachTo: div })
        expect(element.find('.streaming-progress__skipped')).toHaveLength(0)
        expect(element.find('.streaming-progress__skipped-popover')).toHaveLength(0)
    })

    it('should be in info state with only info items', () => {
        const progress: Progress = {
            done: true,
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

        const element = mount(<StreamingProgressSkippedButton progress={progress} />, { attachTo: div })
        expect(element.find('.btn.streaming-progress__skipped')).toHaveLength(1)
        expect(element.find('.btn.streaming-progress__skipped.alert.alert-danger')).toHaveLength(0)
    })

    it('should be in warning state with at least one warning item', () => {
        const progress: Progress = {
            done: true,
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

        const element = mount(<StreamingProgressSkippedButton progress={progress} />, { attachTo: div })
        expect(element.find('.btn.streaming-progress__skipped')).toHaveLength(1)
        expect(element.find('.btn.streaming-progress__skipped--warning')).toHaveLength(1)
    })

    it('should open and close popover when button is clicked', () => {
        const progress: Progress = {
            done: true,
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

        const element = mount(<StreamingProgressSkippedButton progress={progress} />, { attachTo: div })

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
})
