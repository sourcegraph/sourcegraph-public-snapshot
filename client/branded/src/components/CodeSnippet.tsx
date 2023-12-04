import React, { useMemo } from 'react'

import { mdiContentCopy } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'

import { highlightCodeSafe } from '@sourcegraph/common'
import { Button, Code, Icon } from '@sourcegraph/wildcard'

import styles from './CodeSnippet.module.scss'

interface CodeSnippetProps {
    /** The code to be displayed. */
    code: string
    /** Hint to the language, used for syntax-highlighting the code-snippet. */
    language: string
    /** Enable inclusion of a "Copy" button in snippet */
    withCopyButton?: boolean

    className?: string
}

export const CodeSnippet: React.FunctionComponent<React.PropsWithChildren<CodeSnippetProps>> = ({
    code,
    language,
    className,
    withCopyButton = false,
}) => {
    const highlightedInput = useMemo(() => ({ __html: highlightCodeSafe(code, language) }), [code, language])
    return (
        <pre className={classNames('bg-code rounded border p-3 position-relative', className)}>
            {withCopyButton && (
                <Button className={styles.copyButton} onClick={() => copy(code)}>
                    <Icon className="pr-2 pt-2" svgPath={mdiContentCopy} inline={false} aria-label="Copy snippet" />
                </Button>
            )}
            <Code dangerouslySetInnerHTML={highlightedInput} />
        </pre>
    )
}
