import { mdiWrench } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Icon, Link, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import type { CodeIntelIndexerFields } from '../../../../graphql-operations'

interface ConfigurationStateBadgeProps {
    indexer: CodeIntelIndexerFields
    className?: string
}

export const ConfigurationStateBadge: React.FunctionComponent<ConfigurationStateBadgeProps> = ({
    indexer,
    className,
}) => {
    const [ref, truncated, checkTruncation] = useIsTruncated<HTMLAnchorElement>()

    return (
        <Tooltip content={truncated ? indexer.key : null}>
            <Badge
                as={Link}
                to="../index-configuration"
                variant="outlineSecondary"
                key={indexer.key}
                className={classNames('text-muted', className)}
                ref={ref}
                onFocus={checkTruncation}
                onMouseEnter={checkTruncation}
            >
                <Icon svgPath={mdiWrench} aria-hidden={true} className="mr-1 text-primary" />
                Configure {indexer.key}
            </Badge>
        </Tooltip>
    )
}
