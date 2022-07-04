import { forwardRef } from 'react'


import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Icon } from '@sourcegraph/wildcard'

import {
    InsightQueryInput,
    InsightQueryInputProps,
} from '../../../../../../components/form/query-input/InsightQueryInput'

import styles from './CaptureGroupQueryInput.module.scss'
import { mdiRegex } from "@mdi/js";

export interface CaptureGroupQueryInputProps extends Omit<InsightQueryInputProps, 'patternType'> {}

export const CaptureGroupQueryInput = forwardRef<HTMLInputElement, CaptureGroupQueryInputProps>((props, reference) => (
    <InsightQueryInput {...props} ref={reference} patternType={SearchPatternType.regexp}>
        <Button variant="icon" className={styles.regexButton} disabled={true}>
            <Icon
                data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input." svgPath={mdiRegex} inline={false} aria-hidden={true} height={21} width={21}
            />
        </Button>
    </InsightQueryInput>
))
