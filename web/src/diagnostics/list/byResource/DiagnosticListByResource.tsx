import { Diagnostic } from '@sourcegraph/extension-api-types'
import { sortBy } from 'lodash'
import React, { useMemo } from 'react'
import * as sourcegraph from 'sourcegraph'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { parseRepoURI, toPrettyBlobURL } from '../../../../../shared/src/util/url'
import { DiagnosticListItem } from '../DiagnosticListItem'

interface Props extends ExtensionsControllerProps {
    diagnostics: (Diagnostic | sourcegraph.Diagnostic)[]

    className?: string
    listClassName?: string
    itemClassName?: string
}

/**
 * A list of diagnostics, grouped by resource URI.
 */
export const DiagnosticListByResource: React.FunctionComponent<Props> = ({
    diagnostics,
    className = '',
    listClassName = 'list-group',
    itemClassName = 'py-2',
    ...props
}) => {
    const byPath = useMemo(() => {
        const map = new Map<string, (Diagnostic | sourcegraph.Diagnostic)[]>()
        for (const diagnostic of diagnostics) {
            const key = diagnostic.resource.toString()
            const byPath = map.get(key)
            if (byPath) {
                byPath.push(diagnostic)
            } else {
                map.set(key, [diagnostic])
            }
        }
        return sortBy(Array.from(map.entries()), 0)
    }, [diagnostics])

    return byPath.length > 0 ? (
        <div className={className}>
            <ul className={listClassName}>
                {byPath.map(([uri, diagnostics], i) => {
                    const parsedURI = parseRepoURI(uri)
                    return (
                        <li key={i} className="list-group-item">
                            <LinkOrSpan
                                to={toPrettyBlobURL({
                                    ...parsedURI,
                                    rev: parsedURI.rev || parsedURI.commitID!,
                                    filePath: parsedURI.filePath || '',
                                })}
                                className="d-block"
                            >
                                {parsedURI.filePath ? (
                                    <>
                                        <span className="font-weight-normal">
                                            {displayRepoName(parsedURI.repoName)}
                                        </span>{' '}
                                        › {parsedURI.filePath}
                                    </>
                                ) : (
                                    displayRepoName(parsedURI.repoName)
                                )}
                            </LinkOrSpan>
                            <ul className="list-unstyled ml-4 mb-3">
                                {diagnostics.map((diagnostic, j) => (
                                    <li key={j}>
                                        <DiagnosticListItem
                                            {...props}
                                            diagnostic={diagnostic}
                                            className={itemClassName}
                                        />
                                    </li>
                                ))}
                            </ul>
                        </li>
                    )
                })}
            </ul>
        </div>
    ) : null
}
