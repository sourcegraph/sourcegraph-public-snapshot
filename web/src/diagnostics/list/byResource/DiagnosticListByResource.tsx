import { Diagnostic } from '@sourcegraph/extension-api-types'
import { sortBy } from 'lodash'
import React, { useCallback, useMemo, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { parseRepoURI, toPrettyBlobURL } from '../../../../../shared/src/util/url'
import { ThemeProps } from '../../../theme'
import { DiagnosticListItem } from '../DiagnosticListItem'

interface Props extends ExtensionsControllerProps, ThemeProps {
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
    itemClassName = 'pt-2',
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

    const DEFAULT_MAX = 5
    const [max, setMax] = useState<number | undefined>(DEFAULT_MAX)
    const onShowAllClick = useCallback(() => setMax(undefined), [])

    return byPath.length > 0 ? (
        <ul className={`${className} ${listClassName}`}>
            {byPath.slice(0, max).map(([uri, diagnostics], i) => {
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
                                    <span className="font-weight-normal">{displayRepoName(parsedURI.repoName)}</span> ›{' '}
                                    {parsedURI.filePath}
                                </>
                            ) : (
                                displayRepoName(parsedURI.repoName)
                            )}
                        </LinkOrSpan>
                        <ul className="list-unstyled ml-4">
                            {diagnostics.map((diagnostic, j) => (
                                <li key={j}>
                                    <DiagnosticListItem {...props} diagnostic={diagnostic} className={itemClassName} />
                                </li>
                            ))}
                        </ul>
                    </li>
                )
            })}
            {max !== undefined && byPath.length > max && (
                <li className="list-group-item card-footer p-0">
                    <button type="button" className="btn btn-sm btn-link" onClick={onShowAllClick}>
                        Show {byPath.length - max} more
                    </button>
                </li>
            )}
        </ul>
    ) : null
}
