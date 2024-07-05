import React, {useRef, useState} from 'react'

import {
    mdiCheckboxBlankCircleOutline,
    mdiCheckCircleOutline,
    mdiDotsHorizontal,
    mdiGithub,
    mdiOpenInNew,
    mdiPencil,
    mdiRefresh,
    mdiTrashCan
} from '@mdi/js'
import classNames from 'classnames'
import {animated, useSpring} from 'react-spring'

import {convertREMToPX} from '@sourcegraph/shared/src/components/utils/size'
import {
    Alert,
    Button,
    ButtonLink,
    H3,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuList,
    Position,
    Text
} from '@sourcegraph/wildcard'

import {defaultExternalServices} from '../../../components/externalServices/externalServices'
import {AppLogo} from '../../../components/gitHubApps/AppLogo'
import {RemoveGitHubAppModal} from '../../../components/gitHubApps/RemoveGitHubAppModal'
import type {BatchChangesCodeHostFields} from '../../../graphql-operations'

import {useRefreshGitHubApp} from './backend'

import styles from './CommitSigningIntegrationNode.module.scss'
import {useNavigate} from 'react-router-dom';

interface CommitSigningIntegrationNodeProps {
    readOnly: boolean
    node: BatchChangesCodeHostFields
    refetch: () => void
}

export const CommitSigningIntegrationNode: React.FunctionComponent<
    React.PropsWithChildren<CommitSigningIntegrationNodeProps>
> = ({ node, readOnly, refetch }) => {
    const ExternalServiceIcon = defaultExternalServices[node.externalServiceKind].icon
    return (
        <li className={classNames(styles.node, 'list-group-item')}>
            <div
                className={classNames(
                    styles.wrapper,
                    'd-flex justify-content-between align-items-center flex-wrap mb-0'
                )}
            >
                <H3 className="mb-0 mr-2">
                    {node.commitSigningConfiguration ? (
                        <Icon
                            aria-label="This code host has commit signing enabled with a GitHub App."
                            className="text-success"
                            svgPath={mdiCheckCircleOutline}
                        />
                    ) : (
                        <Icon
                            aria-label="This code host does not have a GitHub App connected for commit signing."
                            className="text-danger"
                            svgPath={mdiCheckboxBlankCircleOutline}
                        />
                    )}

                    <Icon className="mx-2" aria-hidden={true} as={ExternalServiceIcon} />
                    {node.externalServiceURL}
                </H3>
                {readOnly ? (
                    <ReadOnlyAppDetails config={node.commitSigningConfiguration} />
                ) : (
                    <AppDetailsControls
                        baseURL={node.externalServiceURL}
                        config={node.commitSigningConfiguration}
                        refetch={refetch}
                    />
                )}
            </div>
        </li>
    )
}

interface AppDetailsControlsProps {
    baseURL: string
    config: BatchChangesCodeHostFields['commitSigningConfiguration']
    refetch: () => void
}

const AppDetailsControls: React.FunctionComponent<AppDetailsControlsProps> = ({ baseURL, config, refetch }) => {
    const [removeModalOpen, setRemoveModalOpen] = useState<boolean>(false)
    const [refreshGitHubApp, { loading, error, data }] = useRefreshGitHubApp()
    const createURL = `/site-admin/batch-changes/github-apps/new?baseURL=${encodeURIComponent(baseURL)}`
    const navigate = useNavigate()

    // const mainButtonRef = useRef<HTMLButtonElement>(null)
    // const handleClick = (event: React.MouseEvent) => {
    //     if (mainButtonRef?.current) {
    //         mainButtonRef.current.click();
    //         alert('clicked')
    //     }
    // };

    return config ? (
        <>
            {removeModalOpen && (
                <RemoveGitHubAppModal onCancel={() => setRemoveModalOpen(false)} afterDelete={refetch} app={config} />
            )}
            <Menu>
                <MenuButton
                    outline={true}
                    aria-label="Repository action"
                    className={styles.menuItems}
                    // onClick={e => e.stopPropagation()}
                    // ref={mainButtonRef}
                >
                    <div className={styles.appDetailsControls} role="button" tabIndex={0}>
                        <AppLogo src={config.logo} name={config.name}
                                 className={classNames(styles.appLogoLarge, 'mr-2')}/>

                        <div className={styles.appDetailsColumn}>
                            <Text size="small" className="font-weight-bold mb-0">
                                {config.name}
                            </Text>
                            <Text size="small" className="text-muted mb-0">
                                AppID: {config.appID}
                            </Text>
                        </div>
                        <div className={styles.appDetailsColumn}>
                            <Icon svgPath={mdiDotsHorizontal} inline={false} aria-hidden={true}/>
                        </div>
                    </div>
                </MenuButton>
                <MenuList position={Position.bottomEnd}>
                    <MenuItem
                        as={Button}
                        onSelect={() => window.open(config?.appURL, '_blank')}
                        className="p-2"
                    >
                        <Icon aria-hidden={true} svgPath={mdiGithub} className="mr-1"/>
                        View on GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true}/>
                    </MenuItem>
                    <MenuDivider/>
                    <MenuItem
                        as={Button}
                        disabled={loading}
                        onSelect={() => refreshGitHubApp({ variables: { gitHubApp: config.id } })}
                        className="p-2"
                    >
                        <Icon aria-hidden={true} svgPath={mdiRefresh} className="mr-1" />
                        Refresh
                    </MenuItem>
                    <MenuItem
                        as={Button}
                        onSelect={() =>
                            navigate(`github-apps/${config.id}`)
                        }
                        className="p-2"
                    >
                        <Icon aria-hidden={true} svgPath={mdiPencil} className="mr-1" />
                        Edit
                    </MenuItem>
                    <MenuItem
                        as={Button}
                        onSelect={() => setRemoveModalOpen(true)}
                        className="p-2"
                    >
                        <Icon aria-hidden={true} svgPath={mdiTrashCan} className="mr-1" />
                        Remove
                    </MenuItem>
                </MenuList>
            </Menu>

            {/* <div className="ml-auto"> */}
            {/*     <AnchorLink to={config.appURL} target="_blank" className="mr-3"> */}
            {/*         <small> */}
            {/*             View In GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} /> */}
            {/*         </small> */}
            {/*     </AnchorLink> */}
            {/*     <Button */}
            {/*         variant="warning" */}
            {/*         className="mr-2" */}
            {/*         size="sm" */}
            {/*         onClick={() => refreshGitHubApp({ variables: { gitHubApp: config.id } })} */}
            {/*     > */}
            {/*         {loading ? <LoadingSpinner inline={true} /> : 'Refresh'} */}
            {/*     </Button> */}
            {/*     <ButtonLink */}
            {/*         className="mr-2" */}
            {/*         aria-label="Edit" */}
            {/*         to={`github-apps/${config.id}`} */}
            {/*         variant="secondary" */}
            {/*         size="sm" */}
            {/*     > */}
            {/*         <Icon aria-hidden={true} svgPath={mdiCogOutline} /> Edit */}
            {/*     </ButtonLink> */}
            {/*     <Button */}
            {/*         aria-label="Remove GitHub App" */}
            {/*         onClick={() => setRemoveModalOpen(true)} */}
            {/*         variant="danger" */}
            {/*         size="sm" */}
            {/*     > */}
            {/*         <Icon aria-hidden={true} svgPath={mdiDelete} /> Remove */}
            {/*     </Button> */}
            {/* </div> */}
            {error && <NodeAlert variant="danger">{error.message}</NodeAlert>}
            {!loading && data && (
                <NodeAlert variant="success">
                    Installations for <span className="font-weight-bold">"{config.name}"</span> successfully refreshed.
                </NodeAlert>
            )}
        </>
    ) : (
        <ButtonLink to={createURL} className="ml-auto text-nowrap" variant="success" as={Link} size="sm">
            Create GitHub App
        </ButtonLink>
    )
}

interface ReadOnlyAppDetailsProps {
    config: BatchChangesCodeHostFields['commitSigningConfiguration']
}

const ReadOnlyAppDetails: React.FunctionComponent<ReadOnlyAppDetailsProps> = ({ config }) =>
    config ? (
        <div className={styles.readonlyAppDetails}>
            <img className={styles.appLogo} src={config.logo} alt="app logo" aria-hidden={true} />
            <Text size="small" className="font-weight-bold m-0">
                {config.name}
            </Text>
        </div>
    ) : (
        <div className={styles.readonlyAppDetails}>
            <Text size="small" className="m-0">
                No App configured
            </Text>
        </div>
    )

// The Alert banner has a 1rem bottom margin
const ONE_REM_IN_PX = convertREMToPX(1)
const APPROX_BANNER_HEIGHT_PX = 40

interface NodeAlertProps {
    variant: 'danger' | 'success'
}

const NodeAlert: React.FunctionComponent<React.PropsWithChildren<NodeAlertProps>> = ({ children, variant }) => {
    const ref = useRef<HTMLDivElement>(null)
    const style = useSpring({
        from: {
            height: '0px',
            opacity: 0,
        },
        to: {
            height: `${(ref.current?.offsetHeight || APPROX_BANNER_HEIGHT_PX) + ONE_REM_IN_PX}px`,
            opacity: 1,
        },
    })

    return (
        <animated.div style={style}>
            {/* Keep this in sync with calculation above: mb-3 = 1rem */}
            <Alert ref={ref} variant={variant} className="mb-3">
                {children}
            </Alert>
        </animated.div>
    )
}
