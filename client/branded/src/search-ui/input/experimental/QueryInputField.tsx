import React, { useMemo } from 'react'

import { Prec } from '@codemirror/state'
import { keymap } from '@codemirror/view'

import { isDefined } from '@sourcegraph/common'
import type { QueryState } from '@sourcegraph/shared/src/search'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

import { changeListener, createDefaultSuggestions } from '../codemirror'
import { CodeMirrorQueryInput } from '../CodeMirrorQueryInput'

interface QueryInputFieldProps
    extends Pick<React.ComponentProps<typeof CodeMirrorQueryInput>, 'patternType' | 'className'> {
    value: QueryState
    onChange: (value: QueryState) => void
    onSubmit?: () => void
}

export const QueryInputField: React.FunctionComponent<React.PropsWithChildren<QueryInputFieldProps>> = ({
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

    const extensions = useMemo(
        () =>
            [
                onSubmit
                    ? Prec.highest(
                          keymap.of([
                              {
                                  key: 'Mod-Enter',
                                  run: () => {
                                      onSubmit()
                                      return true
                                  },
                              },
                          ])
                      )
                    : null,
                changeListener(value => onChange({ query: value })),
                autocompletion,
            ].filter(isDefined),
        [autocompletion, onSubmit, onChange]
    )

    return (
        <CodeMirrorQueryInput
            interpretComments={false}
            value={value.query}
            extensions={extensions}
            patternType={props.patternType}
            className={props.className}
        />
    )
}
