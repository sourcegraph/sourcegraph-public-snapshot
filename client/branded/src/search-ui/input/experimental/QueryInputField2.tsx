import React, { useMemo } from 'react'

import { isDefined } from '@sourcegraph/common'
import type { QueryState } from '@sourcegraph/shared/src/search'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'

import { createDefaultSuggestions } from '../codemirror'

import { CodeMirrorQueryInputWrapper } from './CodeMirrorQueryInputWrapper'

const noopOnSubmit = (): void => {}

interface QueryInputFieldProps
    extends Pick<React.ComponentProps<typeof CodeMirrorQueryInputWrapper>, 'patternType' | 'className'> {
    value: QueryState // TODO(sqs): inherit from parent
    onChange: (value: QueryState) => void
    onSubmit?: () => void
    placeholder?: string
}

export const QueryInputField2: React.FunctionComponent<React.PropsWithChildren<QueryInputFieldProps>> = ({
    value,
    onChange,
    onSubmit,
    ...props
}) => {
    const isSourcegraphDotCom = false // TODO(sqs)

    const autocompletion = useMemo(
        () =>
            createDefaultSuggestions({
                fetchSuggestions: query => fetchStreamSuggestions(query),
                isSourcegraphDotCom,
            }),
        [isSourcegraphDotCom]
    )

    const extensions = useMemo(() => [autocompletion].filter(isDefined), [autocompletion])

    const isLightTheme = useIsLightTheme()

    return (
        <CodeMirrorQueryInputWrapper
            interpretComments={false}
            showHistory={false}
            isLightTheme={isLightTheme}
            queryState={value}
            onChange={onChange}
            onSubmit={onSubmit ?? noopOnSubmit}
            placeholder={props.placeholder || ''}
            extensions={extensions}
            patternType={props.patternType}
            className={props.className}
            visualMode="flush"
        >
            {props.children}
        </CodeMirrorQueryInputWrapper>
    )
}
