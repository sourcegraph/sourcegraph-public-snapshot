import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button } from '@sourcegraph/wildcard'

import styles from './MonacoPreviewLink.module.scss'

export interface MonacoPreviewLinkProps {
    query: string
    patternType: SearchPatternType
    className?: string
}

export const MonacoPreviewLink: React.FunctionComponent<MonacoPreviewLinkProps> = props => {
    const { query, patternType, className } = props
    const queryURL = useMemo(() => `/search?${buildSearchURLQuery(query, patternType, false)}`, [patternType, query])

    return (
        <Button className={classNames(styles.previewLink, className)} to={queryURL} variant="link" as={Link}>
            Preview results <LinkExternalIcon size={18} className={styles.previewLink} />
        </Button>
    )
}
