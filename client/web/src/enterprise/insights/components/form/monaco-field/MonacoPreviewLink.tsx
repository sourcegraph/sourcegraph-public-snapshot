import React, { useMemo } from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Icon, Link } from '@sourcegraph/wildcard'

import styles from './MonacoPreviewLink.module.scss'

export interface MonacoPreviewLinkProps {
    query: string
    patternType: SearchPatternType
    className?: string
}

export const MonacoPreviewLink: React.FunctionComponent<React.PropsWithChildren<MonacoPreviewLinkProps>> = props => {
    const { query, patternType, className } = props
    const queryURL = useMemo(() => `/search?${buildSearchURLQuery(query, patternType, false)}`, [patternType, query])

    return (
        <Button
            className={classNames(styles.previewLink, className)}
            to={queryURL}
            variant="link"
            as={Link}
            target="_blank"
            rel="noopener noreferrer"
        >
            Preview results{' '}
            <Icon
                height={18}
                width={18}
                className={styles.previewLink}
                svgPath={mdiOpenInNew}
                inline={false}
                aria-hidden={true}
            />
        </Button>
    )
}
