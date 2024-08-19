import React, { useEffect, useLayoutEffect, useRef, useState } from 'react'

import type { Meta, StoryFn } from '@storybook/react'
import classNames from 'classnames'
import { noop } from 'rxjs'

import { Popover, PopoverContent, type PopoverOpenEvent, PopoverTail, PopoverTrigger, Position } from '..'
import { BrandedStory } from '../../../stories/BrandedStory'
import { Button } from '../../Button'
import { createRectangle, type Point, Strategy } from '../tether'

import styles from './Popover.story.module.scss'

const config: Meta = {
    title: 'wildcard/Popover',
    component: Popover,
    decorators: [story => <BrandedStory>{() => story()}</BrandedStory>],
    parameters: {
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=954%3A1352',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=954%3A2975',
            },
        ],
    },
}

export default config

export const PositionSettingsGallery: StoryFn = () => {
    const [position, setPosition] = useState(Position.top)

    return (
        <div className={classNames(styles.container, 'd-flex justify-content-center align-items-center')}>
            <div className={styles.positionsContainer}>
                <Popover isOpen={true} onOpenChange={noop}>
                    <PopoverTrigger className={styles.positionsTarget} as="div">
                        Target
                    </PopoverTrigger>

                    <PopoverContent
                        position={position}
                        focusLocked={false}
                        className={classNames(styles.floating, styles.floatingTooltipLike)}
                    >
                        Position {position}
                    </PopoverContent>
                </Popover>

                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerTop,
                        styles.positionMarkerTopStart,
                        { [styles.positionMarkerActive]: position === Position.topStart }
                    )}
                    onClick={() => setPosition(Position.topStart)}
                />
                <button
                    className={classNames(styles.positionMarker, styles.positionMarkerTop, {
                        [styles.positionMarkerActive]: position === Position.top,
                    })}
                    onClick={() => setPosition(Position.top)}
                />
                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerTop,
                        styles.positionMarkerTopEnd,
                        { [styles.positionMarkerActive]: position === Position.topEnd }
                    )}
                    onClick={() => setPosition(Position.topEnd)}
                />

                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerRight,
                        styles.positionMarkerRightStart,
                        { [styles.positionMarkerActive]: position === Position.rightStart }
                    )}
                    onClick={() => setPosition(Position.rightStart)}
                />
                <button
                    className={classNames(styles.positionMarker, styles.positionMarkerRight, {
                        [styles.positionMarkerActive]: position === Position.right,
                    })}
                    onClick={() => setPosition(Position.right)}
                />
                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerRight,
                        styles.positionMarkerRightEnd,
                        { [styles.positionMarkerActive]: position === Position.rightEnd }
                    )}
                    onClick={() => setPosition(Position.rightEnd)}
                />

                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerBottom,
                        styles.positionMarkerBottomStart,
                        { [styles.positionMarkerActive]: position === Position.bottomStart }
                    )}
                    onClick={() => setPosition(Position.bottomStart)}
                />
                <button
                    className={classNames(styles.positionMarker, styles.positionMarkerBottom, {
                        [styles.positionMarkerActive]: position === Position.bottom,
                    })}
                    onClick={() => setPosition(Position.bottom)}
                />
                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerBottom,
                        styles.positionMarkerBottomEnd,
                        { [styles.positionMarkerActive]: position === Position.bottomEnd }
                    )}
                    onClick={() => setPosition(Position.bottomEnd)}
                />

                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerLeft,
                        styles.positionMarkerLeftStart,
                        { [styles.positionMarkerActive]: position === Position.leftStart }
                    )}
                    onClick={() => setPosition(Position.leftStart)}
                />
                <button
                    className={classNames(styles.positionMarker, styles.positionMarkerLeft, {
                        [styles.positionMarkerActive]: position === Position.left,
                    })}
                    onClick={() => setPosition(Position.left)}
                />
                <button
                    className={classNames(
                        styles.positionMarker,
                        styles.positionMarkerLeft,
                        styles.positionMarkerLeftEnd,
                        { [styles.positionMarkerActive]: position === Position.leftEnd }
                    )}
                    onClick={() => setPosition(Position.leftEnd)}
                />
            </div>
        </div>
    )
}

export const StandardExample: StoryFn = () => (
    <ScrollCenterBox title="Root scroll block" className={styles.container}>
        <div className={styles.content}>
            <Popover>
                <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                    Hello
                </PopoverTrigger>

                <PopoverContent position={Position.rightStart} className={styles.floating}>
                    Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                    Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                    career and his mother was a homemaker.[6] In the early years of his life his family moved to Kharkiv
                    in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv National
                    Pedagogical University.
                    <div className="mt-2 d-flex" style={{ gap: 10 }}>
                        <Button variant="secondary">Action 1</Button>
                        <Button variant="secondary">Action 2</Button>
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    </ScrollCenterBox>
)

const TARGET_PADDING = createRectangle(0, 0, 10, 10)

export const TargetPaddingExample: StoryFn = () => (
    <ScrollCenterBox title="Root scroll block" className={styles.container}>
        <div className={styles.content}>
            <Popover>
                <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                    Hello
                </PopoverTrigger>

                <PopoverContent
                    targetPadding={TARGET_PADDING}
                    position={Position.bottomStart}
                    className={styles.floating}
                >
                    Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                    Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                    career and his mother was a homemaker.[6] In the early years of his life his family moved to Kharkiv
                    in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv National
                    Pedagogical University.
                    <div className="mt-2 d-flex" style={{ gap: 10 }}>
                        <Button variant="secondary">Action 1</Button>
                        <Button variant="secondary">Action 2</Button>
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    </ScrollCenterBox>
)

export const AbsoluteStrategyExample: StoryFn = () => (
    <ScrollCenterBox title="Root scroll block" className={styles.container}>
        <div className={styles.content}>
            <Popover>
                <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                    Hello
                </PopoverTrigger>

                <PopoverContent
                    position={Position.rightStart}
                    constrainToScrollParents={true}
                    overflowToScrollParents={true}
                    strategy={Strategy.Absolute}
                    className={styles.floating}
                >
                    Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                    Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                    career and his mother was a homemaker.[6] In the early years of his life his family moved to Kharkiv
                    in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv National
                    Pedagogical University.
                    <div className="mt-2 d-flex" style={{ gap: 10 }}>
                        <Button variant="secondary">Action 1</Button>
                        <Button variant="secondary">Action 2</Button>
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    </ScrollCenterBox>
)

export const WithCustomAnchor: StoryFn = () => {
    const customAnchor = useRef<HTMLDivElement>(null)

    return (
        <ScrollCenterBox title="Root scroll block" className={styles.container}>
            <div className={styles.content}>
                <Popover anchor={customAnchor}>
                    <div ref={customAnchor} className={styles.triggerAnchor}>
                        <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                            Hello
                        </PopoverTrigger>
                    </div>

                    <PopoverContent position={Position.rightStart} className={styles.floating}>
                        Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                        Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                        career and his mother was a homemaker.[6] In the early years of his life his family moved to
                        Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                        National Pedagogical University.
                        <div className="mt-2 d-flex" style={{ gap: 10 }}>
                            <Button variant="secondary">Action 1</Button>
                            <Button variant="secondary">Action 2</Button>
                        </div>
                    </PopoverContent>
                </Popover>
            </div>
        </ScrollCenterBox>
    )
}

enum FSM_STATES {
    Initial = 'Initial',
    PopupOpened = 'PopupOpened',
    FocusedAfterPopupClosed = 'FocusedAfterPopupClosed',
}

enum FSM_ACTIONS {
    TargetFocus,
    TargetBlur,
    PopupClose,
}

const FSM_TRANSITIONS: Record<FSM_STATES, Partial<Record<FSM_ACTIONS, FSM_STATES>>> = {
    [FSM_STATES.Initial]: {
        [FSM_ACTIONS.TargetFocus]: FSM_STATES.PopupOpened,
    },
    [FSM_STATES.PopupOpened]: {
        [FSM_ACTIONS.PopupClose]: FSM_STATES.FocusedAfterPopupClosed,
    },
    [FSM_STATES.FocusedAfterPopupClosed]: {
        [FSM_ACTIONS.TargetBlur]: FSM_STATES.Initial,
    },
}

export const ShowOnFocus: StoryFn = () => {
    const [state, setState] = useState<FSM_STATES>(FSM_STATES.Initial)

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        const nextStateAfterClose = FSM_TRANSITIONS[state][FSM_ACTIONS.PopupClose]

        if (!event.isOpen && nextStateAfterClose) {
            setState(nextStateAfterClose)
        }
    }

    const handleTargetFocus = () => {
        const nextState = FSM_TRANSITIONS[state][FSM_ACTIONS.TargetFocus]

        if (nextState) {
            setState(nextState)
        }
    }

    const handleTargetBlur = () => {
        const nextState = FSM_TRANSITIONS[state][FSM_ACTIONS.TargetBlur]

        if (nextState) {
            setState(nextState)
        }
    }

    const open = state === FSM_STATES.PopupOpened

    return (
        <ScrollCenterBox title="Root scroll block" className={styles.container}>
            <div className={styles.content}>
                <Popover isOpen={open} onOpenChange={handleOpenChange}>
                    <PopoverTrigger
                        as={Button}
                        variant="secondary"
                        className={styles.target}
                        onFocus={handleTargetFocus}
                        onBlur={handleTargetBlur}
                    >
                        Target
                    </PopoverTrigger>

                    <PopoverContent position={Position.rightStart} className={styles.floating}>
                        Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                        Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                        career and his mother was a homemaker.[6] In the early years of his life his family moved to
                        Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                        National Pedagogical University.
                        <div className="mt-2 d-flex" style={{ gap: 10 }}>
                            <Button variant="secondary">Action 1</Button>
                            <Button variant="secondary">Action 2</Button>
                        </div>
                    </PopoverContent>
                </Popover>
            </div>
        </ScrollCenterBox>
    )
}

export const WithControlledState: StoryFn = () => {
    const [open, setOpen] = useState<boolean>(false)
    const handleOpenChange = (event: PopoverOpenEvent): void => {
        setOpen(event.isOpen)
        console.log('REASON', event.reason)
    }

    return (
        <ScrollCenterBox title="Root scroll block" className={styles.container}>
            <div className={styles.content}>
                <Button variant="primary" onClick={() => setOpen(true)}>
                    Open popover
                </Button>

                <Popover isOpen={open} onOpenChange={handleOpenChange}>
                    <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                        Target
                    </PopoverTrigger>

                    <PopoverContent position={Position.rightStart} className={styles.floating}>
                        Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                        Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                        career and his mother was a homemaker.[6] In the early years of his life his family moved to
                        Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                        National Pedagogical University.
                        <div className="mt-2 d-flex" style={{ gap: 10 }}>
                            <Button variant="secondary">Action 1</Button>
                            <Button variant="secondary">Action 2</Button>
                        </div>
                    </PopoverContent>
                </Popover>
            </div>
        </ScrollCenterBox>
    )
}

export const WithNestedScrollParents: StoryFn = (args = {}) => {
    const constrainToScrollParents = args.constrainToScrollParents

    return (
        <ScrollCenterBox title="Root scroll block" className={styles.root}>
            <div className={styles.spreadContentBlock}>
                <ScrollCenterBox
                    title="Sub scroll block (see controls panel for rendering tooltip outside of the scroll container"
                    className={classNames(styles.container, styles.containerAsSubRoot)}
                >
                    <div className={styles.content}>
                        <Popover>
                            <div className={styles.triggerAnchor}>
                                <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                                    Hello
                                </PopoverTrigger>
                            </div>

                            <PopoverContent
                                constrainToScrollParents={constrainToScrollParents}
                                position={Position.rightStart}
                                className={styles.floating}
                            >
                                Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky
                                Oblast (now Nizhny Novgorod Oblast). Limonov's father—then in the military service – was
                                in a state security career and his mother was a homemaker.[6] In the early years of his
                                life his family moved to Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied
                                at the H.S. Skovoroda Kharkiv National Pedagogical University.
                                <div className="mt-2 d-flex" style={{ gap: 10 }}>
                                    <Button variant="secondary">Action 1</Button>
                                    <Button variant="secondary">Action 2</Button>
                                </div>
                            </PopoverContent>
                        </Popover>
                    </div>
                </ScrollCenterBox>
            </div>
        </ScrollCenterBox>
    )
}
WithNestedScrollParents.argTypes = {
    constrainToScrollParents: {
        control: { type: 'boolean' },
    },
}
WithNestedScrollParents.args = {
    constrainToScrollParents: true,
}

export const WithVirtualTarget: StoryFn = () => {
    const [virtualElement, setVirtualElement] = useState<Point | null>(null)
    const activeZoneReference = useRef<HTMLDivElement>(null)

    useEffect(() => {
        const element = activeZoneReference.current

        if (!element) {
            return
        }

        function handleMove(event: PointerEvent): void {
            setVirtualElement({
                x: event.clientX,
                y: event.clientY,
            })
        }

        element.addEventListener('pointermove', handleMove)
        element.addEventListener('pointerleave', () => setVirtualElement(null))

        return () => {
            element.removeEventListener('pointermove', handleMove)
        }
    }, [])

    return (
        <div ref={activeZoneReference} className={styles.container}>
            <span className="m-auto">Hover me</span>
            {virtualElement && (
                <PopoverContent
                    isOpen={true}
                    pin={virtualElement}
                    position={Position.rightStart}
                    className={classNames(styles.floating, styles.floatingWithNonEvents)}
                >
                    Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                    Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                    career and his mother was a homemaker.[6] In the early years of his life his family moved to Kharkiv
                    in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv National
                    Pedagogical University.
                </PopoverContent>
            )}
        </div>
    )
}

export const WithTail: StoryFn = (args = {}) => (
    <ScrollCenterBox title="Root scroll block" className={styles.container}>
        <div className={styles.content}>
            <Popover>
                <PopoverTrigger as={Button} variant="secondary" className={styles.target}>
                    Hello
                </PopoverTrigger>

                <PopoverContent position={Position.rightStart} className={styles.floating}>
                    Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                    Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                    career and his mother was a homemaker.[6] In the early years of his life his family moved to Kharkiv
                    in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv National
                    Pedagogical University.
                    <div className="mt-2 d-flex" style={{ gap: 10 }}>
                        <Button variant="secondary">Action 1</Button>
                        <Button variant="secondary">Action 2</Button>
                    </div>
                </PopoverContent>

                <PopoverTail size={args.size} />
            </Popover>
        </div>
    </ScrollCenterBox>
)

WithTail.argTypes = {
    size: {
        control: 'radio',
        options: ['sm', 'md', 'lg'],
    },
}
WithTail.args = {
    size: 'sm',
}

interface ScrollCenterBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    title: string
}

const ScrollCenterBox: React.FunctionComponent<React.PropsWithChildren<ScrollCenterBoxProps>> = props => {
    const { children, title, ...otherProps } = props
    const reference = useRef<HTMLDivElement>(null)

    useLayoutEffect(() => {
        if (!reference.current) {
            return
        }

        const { width, height } = reference.current.getBoundingClientRect()

        reference.current.scrollLeft = (reference.current.scrollWidth - width) / 2
        reference.current.scrollTop = (reference.current.scrollHeight - height) / 2
    }, [])

    return (
        <div {...otherProps} ref={reference} className={classNames(otherProps.className, styles.scrollbox)}>
            <span className={styles.scrollboxTitle}> {title} </span>
            {children}
        </div>
    )
}
