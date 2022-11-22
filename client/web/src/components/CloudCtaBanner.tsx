import React from 'react'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Text } from '@sourcegraph/wildcard'

import styles from './CloudCtaBanner.module.scss'

export const CloudCtaBanner: React.FunctionComponent<
    React.PropsWithChildren<{ outlined?: boolean; className?: string; children: React.ReactNode }>
> = ({ outlined = false, className, children }) => (
    <section
        className={classNames(
            outlined ? styles.containerOutline : styles.primaryBg,
            className,
            'd-flex justify-content-center'
        )}
    >
        <Icon className="mr-2 text-merged" size={outlined ? 'sm' : 'md'} aria-hidden={true} svgPath={mdiArrowRight} />

        <Text size={outlined ? 'small' : 'base'} className="my-auto">
            {children}
        </Text>
    </section>
)
