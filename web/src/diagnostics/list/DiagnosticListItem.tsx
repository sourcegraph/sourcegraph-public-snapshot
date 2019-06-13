import { Diagnostic } from '@sourcegraph/extension-api-types'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuRightIcon from 'mdi-react/MenuRightIcon'
import React, { useCallback, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { CodeExcerpt } from '../../../../shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { parseRepoURI } from '../../../../shared/src/util/url'
import { fetchHighlightedFileLines } from '../../repo/backend'
import { ThemeProps } from '../../theme'
import { DiagnosticMessageWithIcon } from '../components/DiagnosticMessageWithIcon'

interface Props extends ExtensionsControllerProps, ThemeProps {
    diagnostic: Diagnostic | sourcegraph.Diagnostic

    className?: string
}

const CONTEXT_LINES = 2

export const DiagnosticListItem: React.FunctionComponent<Props> = ({ diagnostic, className = '', ...props }) => {
    const uri = parseRepoURI(diagnostic.resource.toString())
    const [isExcerptVisible, setIsExcerptVisible] = useState(false)
    const toggleIsExcerptVisible = useCallback(() => setIsExcerptVisible(!isExcerptVisible), [isExcerptVisible])
    const ToggleIcon: React.ComponentType<{ className?: string }> = isExcerptVisible ? MenuDownIcon : MenuRightIcon
    return (
        <div className={className}>
            <div onClick={toggleIsExcerptVisible} className="d-flex align-items-center user-select-none">
                <ToggleIcon className="icon-inline h4 mb-0" />
                <DiagnosticMessageWithIcon diagnostic={diagnostic} />
            </div>
            {isExcerptVisible && (
                <CodeExcerpt
                    {...props}
                    repoName={uri.repoName}
                    commitID={uri.commitID!}
                    filePath={uri.filePath!}
                    context={Math.ceil((diagnostic.range.end.line - diagnostic.range.start.line) / 2) + CONTEXT_LINES}
                    highlightRanges={[
                        {
                            // TODO!(sqs): remove '!' non-null assertions
                            line: Math.ceil((diagnostic.range.start.line + diagnostic.range.end.line) / 2),
                            character: diagnostic.range.start.character,
                            highlightLength: diagnostic.range.end.character - diagnostic.range.start.character,
                        },
                    ]}
                    className="w-100 h-100 overflow-auto p-2 d-block"
                    fetchHighlightedFileLines={fetchHighlightedFileLines}
                />
            )}
        </div>
    )
}
