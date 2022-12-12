import React, { useState } from 'react'

import { mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import {
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import styles from './RequestLink.module.scss'

export interface RequestLinkProps {
    indexerName: string
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
}

export const RequestLink: React.FunctionComponent<React.PropsWithChildren<RequestLinkProps>> = ({
    indexerName,
    useRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery,
}) => {
    const language = indexerName.startsWith('lsif-') ? indexerName.slice('lsif-'.length) : indexerName

    const {
        data,
        loading: loadingSupport,
        error,
    } = useRequestedLanguageSupportQuery({
        variables: {},
    })

    const [requested, setRequested] = useState(false)

    const [requestSupport, { loading: requesting, error: requestError }] = useRequestLanguageSupportQuery({
        variables: { language },
        onCompleted: () => setRequested(true),
    })

    return loadingSupport || requesting ? (
        <div className="px-2 py-1">
            <LoadingSpinner />
        </div>
    ) : error ? (
        <div className="px-2 py-1">
            <ErrorAlert prefix="Error loading repository summary" error={error} />
        </div>
    ) : requestError ? (
        <div className="px-2 py-1">
            <ErrorAlert prefix="Error requesting language support" error={requestError} />
        </div>
    ) : data ? (
        data.languages.includes(language) || requested ? (
            <span className="text-muted">
                Received your request{' '}
                <Tooltip content="Requests are documented and contribute to our precise support roadmap">
                    <Icon
                        aria-label="Requests are documented and contribute to our precise support roadmap"
                        svgPath={mdiInformationOutline}
                        inline={false}
                        height={16}
                        width={16}
                    />
                </Tooltip>
            </span>
        ) : (
            <Button variant="link" className={classNames('m-0 p-0', styles.languageRequest)} onClick={requestSupport}>
                I want precise support!
            </Button>
        )
    ) : null
}
