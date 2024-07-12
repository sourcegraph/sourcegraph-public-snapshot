import React, { useRef, useState } from 'react'

import { mdiConnection, mdiDotsHorizontal, mdiGithub, mdiOpenInNew, mdiPencil, mdiRefresh, mdiTrashCan } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'
import { animated, useSpring } from 'react-spring'

import { useLazyQuery } from '@sourcegraph/http-client'
import { convertREMToPX } from '@sourcegraph/shared/src/components/utils/size'
import {
    Alert,
    Button,
    ButtonLink,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuList,
    Position,
    Text,
} from '@sourcegraph/wildcard'

import { AppLogo } from '../../../components/gitHubApps/AppLogo'
import { RemoveGitHubAppModal } from '../../../components/gitHubApps/RemoveGitHubAppModal'
import {
    type CheckBatchChangesCredentialResult,
    type CheckBatchChangesCredentialVariables,
    type GitHubAppByIDFields,
    GitHubAppKind,
} from '../../../graphql-operations'

import { useRefreshGitHubApp, CHECK_BATCH_CHANGES_CREDENTIAL } from './backend'

import styles from './CommitSigningIntegrationNode.module.scss'

interface GitHubAppControlsProps {
    baseURL: string
    config: Pick<GitHubAppByIDFields, 'id' | 'name' | 'appURL' | 'logo' | 'appID'> | null
    refetch: () => void
    gitHubAppKind?: GitHubAppKind

    /**
     * If the GitHub app is used as a credential, we need the credential ID, so we can use that
     * to check if the GitHub app connection is still active.
     * */
    credentialID?: string
}

export const GitHubAppControls: React.FunctionComponent<GitHubAppControlsProps> = ({
    baseURL,
    config,
    refetch,
    gitHubAppKind,
    credentialID,
}) => {
    const [removeModalOpen, setRemoveModalOpen] = useState<boolean>(false)
    const [refreshGitHubApp, { loading, error, data }] = useRefreshGitHubApp()
    const createURL = `/site-admin/batch-changes/github-apps/new?baseURL=${encodeURIComponent(baseURL)}`
    const navigate = useNavigate()

    const [checkCredFn, { data: checkCredResult, loading: checkCredLoading, error: checkCredErr }] = useLazyQuery<
        CheckBatchChangesCredentialResult,
        CheckBatchChangesCredentialVariables
    >(CHECK_BATCH_CHANGES_CREDENTIAL, {})

    return config ? (
        <div>
            {removeModalOpen && (
                <RemoveGitHubAppModal onCancel={() => setRemoveModalOpen(false)} afterDelete={refetch} app={config} />
            )}
            <Menu>
                <MenuButton outline={true} aria-label="Repository action">
                    <div className={styles.gitHubAppControls} role="button" tabIndex={0}>
                        <AppLogo
                            src={config.logo}
                            name={config.name}
                            className={classNames(styles.appLogoLarge, 'mr-2')}
                        />

                        <div className={styles.appDetailsColumn}>
                            <Text size="small" className="font-weight-bold mb-0">
                                {config.name}
                            </Text>
                            <Text size="small" className="text-muted mb-0">
                                AppID: {config.appID}
                            </Text>
                        </div>
                        <div className={styles.appDetailsColumn}>
                            <Icon svgPath={mdiDotsHorizontal} inline={false} aria-hidden={true} />
                        </div>
                    </div>
                </MenuButton>
                <MenuList position={Position.bottomEnd} className={styles.menuList}>
                    <MenuItem as={Button} onSelect={() => window.open(config?.appURL, '_blank')} className="p-2">
                        <Icon aria-hidden={true} svgPath={mdiGithub} className="mr-1" />
                        View on GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} />
                    </MenuItem>
                    <MenuDivider />
                    {credentialID && (
                        <MenuItem
                            as={Button}
                            disabled={checkCredLoading}
                            onSelect={() => checkCredFn({ variables: { id: credentialID } })}
                            className="p-2"
                        >
                            <Icon aria-hidden={true} svgPath={mdiConnection} className="mr-1" />
                            Check
                        </MenuItem>
                    )}
                    <MenuItem
                        as={Button}
                        disabled={loading}
                        onSelect={() => refreshGitHubApp({ variables: { gitHubApp: config.id } })}
                        className="p-2"
                    >
                        <Icon aria-hidden={true} svgPath={mdiRefresh} className="mr-1" />
                        Refresh
                    </MenuItem>
                    {
                        // Once we have a GitHubApps page for users, we can drop this check. Currently, we only have that for site-admins.
                        gitHubAppKind !== GitHubAppKind.USER_CREDENTIAL && (
                            <MenuItem as={Button} onSelect={() => navigate(`github-apps/${config.id}`)} className="p-2">
                                <Icon aria-hidden={true} svgPath={mdiPencil} className="mr-1" />
                                Edit
                            </MenuItem>
                        )
                    }
                    <MenuItem as={Button} onSelect={() => setRemoveModalOpen(true)} className="p-2">
                        <Icon aria-hidden={true} svgPath={mdiTrashCan} className="mr-1" />
                        Remove
                    </MenuItem>
                </MenuList>
            </Menu>
            {error && <NodeAlert variant="danger">{error.message}</NodeAlert>}
            {!loading && data && (
                <NodeAlert variant="success">
                    Installations for <span className="font-weight-bold">"{config.name}"</span> successfully refreshed.
                </NodeAlert>
            )}
            {!checkCredLoading && (checkCredResult || checkCredErr) && (
                <>
                    <br />
                    <NodeAlert variant={checkCredErr ? 'danger' : 'success'}>
                        <span className="font-weight-bold">{config.name}</span> is {checkCredErr ? 'not' : ''}{' '}
                        accessible.
                    </NodeAlert>
                </>
            )}
        </div>
    ) : (
        <ButtonLink to={createURL} className="ml-auto text-nowrap" variant="success" as={Link} size="sm">
            Create GitHub App
        </ButtonLink>
    )
}

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
