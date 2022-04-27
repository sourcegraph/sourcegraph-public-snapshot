import { forwardRef } from 'react'

import RegexIcon from 'mdi-react/RegexIcon'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Button } from '@sourcegraph/wildcard'

import {
    InsightQueryInput,
    InsightQueryInputProps,
} from '../../../../../../components/form/query-input/InsightQueryInput'

import styles from './CaptureGroupQueryInput.module.scss'

export interface CaptureGroupQueryInputProps extends Omit<InsightQueryInputProps, 'patternType'> {}

export const CaptureGroupQueryInput = forwardRef<HTMLInputElement, CaptureGroupQueryInputProps>((props, reference) => (
    <InsightQueryInput {...props} ref={reference} patternType={SearchPatternType.regexp}>
        <Button variant="icon" className={styles.regexButton} disabled={true}>
            <RegexIcon
                size={21}
                data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
            />
        </Button>
    </InsightQueryInput>
))
