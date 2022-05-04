import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { CodeIntelligenceBadgeProps as DefaultRepositoryMenuProps } from '../../../../codeintel/CodeIntelligenceBadge'
import {
    massageIndexerSupportMetadata,
    useCodeIntelStatus as defaultUseCodeIntelStatus,
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import { UserFacingCodeIntelligenceBadgeContent } from './UserFacingCodeIntelligenceBadgeContent'

interface CodeIntelligenceBadgeContentStorybookProps {
    isStorybook?: boolean
    now?: () => Date
    useCodeIntelStatus?: typeof defaultUseCodeIntelStatus
    useRequestedLanguageSupportQuery?: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery?: typeof defaultUseRequestLanguageSupportQuery
}

export interface CodeIntelligenceBadgeContentProps
    extends DefaultRepositoryMenuProps,
        CodeIntelligenceBadgeContentStorybookProps {}

export const CodeIntelligenceBadgeContent: React.FunctionComponent<
    React.PropsWithChildren<CodeIntelligenceBadgeContentProps>
> = props => {
    const { data, loading, error } = defaultUseCodeIntelStatus({
        variables: {
            repository: props.repoName,
            commit: props.revision,
            path: props.filePath,
        },
    })

    const indexerSupportMetadata = data && massageIndexerSupportMetadata(data)

    return loading ? (
        <div className="px-2 py-1">
            <LoadingSpinner />
        </div>
    ) : error ? (
        <div className="px-2 py-1">
            <ErrorAlert prefix="Error loading repository summary" error={error} />
        </div>
    ) : data && indexerSupportMetadata ? (
        <UserFacingCodeIntelligenceBadgeContent
            repoName={props.repoName}
            indexerSupportMetadata={indexerSupportMetadata}
            useRequestedLanguageSupportQuery={defaultUseRequestedLanguageSupportQuery}
            useRequestLanguageSupportQuery={defaultUseRequestLanguageSupportQuery}
            settingsCascade={props.settingsCascade}
        />
    ) : null
}
