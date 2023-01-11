import { forwardRef } from 'react'

import { mdiRegex } from '@mdi/js'
import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import {
    InsightQueryInput,
    InsightQueryInputProps,
} from '../../../../../../components/form/query-input/InsightQueryInput'

import styles from './CaptureGroupQueryInput.module.scss'

export interface CaptureGroupQueryInputProps extends Omit<InsightQueryInputProps, 'patternType'> {}

export const CaptureGroupQueryInput = forwardRef<HTMLInputElement, CaptureGroupQueryInputProps>((props, reference) => (
    <InsightQueryInput
        {...props}
        ref={reference}
        patternType={SearchPatternType.regexp}
        className={classNames(props.className, styles.input)}
    >
        <Tooltip content="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input.">
            <Button variant="icon" className={styles.regexButton} disabled={true}>
                <Icon svgPath={mdiRegex} inline={false} height={21} width={21} aria-hidden={true} />
            </Button>
        </Tooltip>
    </InsightQueryInput>
))
