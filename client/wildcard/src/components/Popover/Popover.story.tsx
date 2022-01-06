import { Meta } from '@storybook/react'
import classNames from 'classnames'
import React, { useLayoutEffect, useRef, useState } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { Button } from '@sourcegraph/wildcard'

import { Popover } from './Popover'
import styles from './Popover.story.module.scss'

export default {
    title: 'wildcard/Popover',
    decorators: [story => <BrandedStory styles={webStyles}>{() => story()}</BrandedStory>],
} as Meta

export const StandardExample = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <ScrollCenterBox className={styles.container}>
            <div className={styles.content}>
                <Button variant="secondary" className={styles.target} ref={setButtonElement}>
                    Hello
                </Button>

                {buttonElement && (
                    <Popover
                        className={styles.floating}
                        strategy="absolute"
                        placement="right-start"
                        target={buttonElement}
                    >
                        Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                        Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                        career and his mother was a homemaker.[6] In the early years of his life his family moved to
                        Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                        National Pedagogical University.
                    </Popover>
                )}
            </div>
        </ScrollCenterBox>
    )
}

export const WithNestedScrollParents = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <ScrollCenterBox className={styles.root}>
            <ScrollCenterBox className={classNames(styles.container, styles.containerAsSubRoot)}>
                <div className={styles.content}>
                    <Button variant="secondary" className={styles.target} ref={setButtonElement}>
                        Hello
                    </Button>

                    {buttonElement && (
                        <Popover
                            className={styles.floating}
                            strategy="absolute"
                            placement="right-start"
                            target={buttonElement}
                        >
                            Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast
                            (now Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state
                            security career and his mother was a homemaker.[6] In the early years of his life his family
                            moved to Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S.
                            Skovoroda Kharkiv National Pedagogical University.
                        </Popover>
                    )}
                </div>
            </ScrollCenterBox>
        </ScrollCenterBox>
    )
}

export const WithScrollFloatingElement = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <ScrollCenterBox className={styles.container}>
            <div className={styles.content}>
                <Button variant="secondary" className={styles.target} ref={setButtonElement}>
                    Hello
                </Button>

                {buttonElement && (
                    <Popover
                        className={classNames(styles.floating, styles.floatingWithScroll)}
                        strategy="absolute"
                        placement="right-start"
                        target={buttonElement}
                    >
                        Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast (now
                        Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state security
                        career and his mother was a homemaker.[6] In the early years of his life his family moved to
                        Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S. Skovoroda Kharkiv
                        National Pedagogical University.
                    </Popover>
                )}
            </div>
        </ScrollCenterBox>
    )
}

export const WithFixedStrategy = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <ScrollCenterBox className={styles.root}>
            <ScrollCenterBox className={classNames(styles.container, styles.containerAsSubRoot)}>
                <ScrollCenterBox className={styles.content}>
                    <Button variant="secondary" className={styles.target} ref={setButtonElement}>
                        Hello
                    </Button>

                    {buttonElement && (
                        <Popover className={styles.floating} strategy="fixed" placement="right" target={buttonElement}>
                            Limonov was born in the Soviet Union, in Dzerzhinsk, an industrial town in the Gorky Oblast
                            (now Nizhny Novgorod Oblast). Limonov's father—then in the military service – was in a state
                            security career and his mother was a homemaker.[6] In the early years of his life his family
                            moved to Kharkiv in the Ukrainian SSR, where Limonov grew up. He studied at the H.S.
                            Skovoroda Kharkiv National Pedagogical University.
                        </Popover>
                    )}
                </ScrollCenterBox>
            </ScrollCenterBox>
        </ScrollCenterBox>
    )
}

export const ScrollCenterBox: React.FunctionComponent<React.HTMLAttributes<HTMLDivElement>> = props => {
    const reference = useRef<HTMLDivElement>(null)

    useLayoutEffect(() => {
        if (!reference.current) {
            return
        }

        const { width, height } = reference.current.getBoundingClientRect()

        reference.current.scrollLeft = (reference.current.scrollWidth - width) / 2
        reference.current.scrollTop = (reference.current.scrollHeight - height) / 2
    }, [])

    return <div {...props} ref={reference} />
}
