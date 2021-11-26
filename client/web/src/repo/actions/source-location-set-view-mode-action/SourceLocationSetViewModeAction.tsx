import classNames from 'classnames'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import React, { useRef } from 'react'
import { Link } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'

import { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Menu, MenuButton, MenuPopover, MenuItems, MenuHeader, MenuLink, MenuDivider } from '@sourcegraph/wildcard'

import { ComponentIcon } from '../../../enterprise/catalog/components/ComponentIcon'
import { ComponentTitleWithIconAndKind } from '../../../enterprise/catalog/contributions/tree/SourceLocationSetTitle'
import { TreeOrComponentViewOptionsProps } from '../../../enterprise/catalog/contributions/tree/TreeOrComponent'
import { positionBottomRight } from '../../../enterprise/insights/components/context-menu/utils'
import {
    SourceLocationSetViewModeInfoResult,
    SourceLocationSetViewModeInfoVariables,
    RepositoryFields,
} from '../../../graphql-operations'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
import { RepoHeaderContext } from '../../RepoHeader'

import styles from './SourceLocationSetViewModeAction.module.scss'

// TODO(sqs): LICENSE move to enterprise/

// TODO(sqs): should this show up when there is no repository rev?

interface Props extends Partial<RevisionSpec>, Partial<FileSpec> {
    repo: Pick<RepositoryFields, 'id' | 'name'>

    actionType?: 'nav' | 'dropdown'
}

const SOURCE_LOCATION_SET_VIEW_MODE_INFO = gql`
    query SourceLocationSetViewModeInfo($repository: ID!, $path: String!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                id
                components(path: $path, primary: true, recursive: false) {
                    __typename
                    id
                    name
                    kind
                    description
                    url
                }
            }
        }
    }
`

export const SourceLocationSetViewModeAction: React.FunctionComponent<Props & RepoHeaderContext> = props => {
    const { data, error, loading } = useQuery<
        SourceLocationSetViewModeInfoResult,
        SourceLocationSetViewModeInfoVariables
    >(SOURCE_LOCATION_SET_VIEW_MODE_INFO, {
        variables: { repository: props.repo.id, path: props.filePath || '' },
        fetchPolicy: 'cache-first',
    })
    if (error) {
        throw error
    }
    if (loading && !data) {
        return null
    }

    const components = (data && data.node?.__typename === 'Repository' && data.node.components) ?? null
    if (!components || components.length === 0) {
        return null
    }

    const component = components[0]

    return null

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionButtonLink to={component.url} className="btn" file={true}>
                <ComponentIcon component={component} className="icon-inline mr-1" /> {component.name}
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <ComponentActionPopoverButton
            component={component}
            className={styles.wrapper}
            buttonClassName={classNames('btn btn-icon small border border-secondary px-2', styles.btn)}
        />
    )
}

type ComponentFields = Extract<
    SourceLocationSetViewModeInfoResult['node'],
    { __typename: 'Repository' }
>['components'][number]

export const ComponentActionPopoverButton: React.FunctionComponent<
    {
        component: ComponentFields
        buttonClassName?: string
    } & Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode' | 'treeOrComponentViewModeURL'>
> = ({ component, buttonClassName, treeOrComponentViewMode, treeOrComponentViewModeURL }) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)

    return (
        <Menu>
            <MenuButton
                variant="secondary"
                outline={true}
                className={classNames(
                    'py-1 px-2',
                    styles.btn,
                    treeOrComponentViewMode === 'auto' ? styles.btnViewModeComponent : styles.btnViewModeTree,
                    buttonClassName
                )}
                ref={targetButtonReference}
            >
                <ComponentTitleWithIconAndKind component={component} strong={treeOrComponentViewMode === 'auto'} />
            </MenuButton>
            <MenuPopover position={positionBottomRight}>
                <SourceLocationSetViewModeActionMenuItems
                    treeOrComponentViewMode={treeOrComponentViewMode}
                    treeOrComponentViewModeURL={treeOrComponentViewModeURL}
                />
            </MenuPopover>
        </Menu>
    )
}

export const SourceLocationSetViewModeActionMenuItems: React.FunctionComponent<
    Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode' | 'treeOrComponentViewModeURL'>
> = ({ treeOrComponentViewMode, treeOrComponentViewModeURL }) => {
    const checkIcon = <CheckBoldIcon className="icon-inline" />
    const noCheckIcon = <CheckBoldIcon className="icon-inline invisible" />

    return (
        <MenuItems>
            <MenuHeader>View as...</MenuHeader>
            <MenuDivider />
            <MenuLink as={Link} to={treeOrComponentViewModeURL.auto}>
                {treeOrComponentViewMode === 'auto' ? checkIcon : noCheckIcon} Component
            </MenuLink>
            <MenuLink as={Link} to={treeOrComponentViewModeURL.tree}>
                {treeOrComponentViewMode === 'tree' ? checkIcon : noCheckIcon} Tree
            </MenuLink>
        </MenuItems>
    )
}
