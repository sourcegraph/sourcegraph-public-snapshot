import { Meta } from '@storybook/react';
import classNames from 'classnames';
import React, { useEffect, useLayoutEffect, useRef, useState } from 'react';

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory';
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss';

import { Button } from '../../../Button';
import { Point, Position } from '../../tether';

import { FloatingPanel } from '../FloatingPanel';
import styles from './FloatingPanel.story.module.scss';

export default {
    title: 'wildcard/Popover/Floating-panel',
    decorators: [story => <BrandedStory styles={webStyles}>{() => story()}</BrandedStory>],
} as Meta

export const StandardExample = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <ScrollCenterBox title='Main scroll sandbox' className={styles.container}>
            <div className={styles.content}>
                <Button variant="secondary" className={styles.target} ref={setButtonElement}>
                    Hello
                </Button>

                {buttonElement && (
                    <FloatingPanel
                        target={buttonElement}
                        position={Position.rightTop}
                        className={styles.floating}
                    >
                        Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                        Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                        career and his mother was a homemaker.[6] In the early years of his life his family moved to
                        Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                        National Pedagogical University.
                    </FloatingPanel>
                )}
            </div>
        </ScrollCenterBox>
    )
}

export const WithNestedScrollParents = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <ScrollCenterBox title='Root scroll area' className={styles.root}>
            <ScrollCenterBox title='Sub scroll area' className={classNames(styles.container, styles.containerAsSubRoot)}>
                <div className={styles.content}>
                    <Button variant="secondary" className={styles.target} ref={setButtonElement}>
                        Hello
                    </Button>

                    {buttonElement && (
                        <FloatingPanel
                            target={buttonElement}
                            position={Position.rightTop}
                            className={styles.floating}
                        >
                            Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast
                            (now Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state
                            security career and his mother was a homemaker.[6] In the early years of his life his family
                            moved to Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S.
                            Skovoroda Kharkiv National Pedagogical University.
                        </FloatingPanel>
                    )}
                </div>
            </ScrollCenterBox>
        </ScrollCenterBox>
    )
}

export const WithVirtualTarget = () => {
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
            Hover me
            {virtualElement && (
                <FloatingPanel
                    target={null}
                    pin={virtualElement}
                    position={Position.rightTop}
                    className={classNames(styles.floating, styles.floatingWithNonEvents)}
                >
                    Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                    Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                    career and his mother was a homemaker.[6] In the early years of his life his family moved to
                    Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                    National Pedagogical University.
                </FloatingPanel>
            )}
        </div>
    )
}

interface ScrollCenterBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    title: string
}

const ScrollCenterBox: React.FunctionComponent<ScrollCenterBoxProps> = props => {
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
            <span className={styles.scrollboxTitle}> { title } </span>
            { children }
        </div>
    )
}
