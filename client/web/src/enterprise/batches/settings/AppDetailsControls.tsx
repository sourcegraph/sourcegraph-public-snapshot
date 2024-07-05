import type {BatchChangesCodeHostFields} from '../../../graphql-operations';
import React, {useRef, useState} from 'react';
import {useRefreshGitHubApp} from './backend';
import {useNavigate} from 'react-router-dom';
import {RemoveGitHubAppModal} from '../../../components/gitHubApps/RemoveGitHubAppModal';
import {
    Alert,
    Button,
    ButtonLink,
    Icon, Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuList,
    Position,
    Text
} from '@sourcegraph/wildcard';
import styles from './CommitSigningIntegrationNode.module.scss';
import {AppLogo} from '../../../components/gitHubApps/AppLogo';
import classNames from 'classnames';
import {mdiDotsHorizontal, mdiGithub, mdiOpenInNew, mdiPencil, mdiRefresh, mdiTrashCan} from '@mdi/js';
import {convertREMToPX} from '@sourcegraph/shared/src/components/utils/size';
import {animated, useSpring} from 'react-spring';

interface AppDetailsControlsProps {
    baseURL: string
    config: BatchChangesCodeHostFields['commitSigningConfiguration']
    refetch: () => void
}

export const AppDetailsControls: React.FunctionComponent<AppDetailsControlsProps> = ({ baseURL, config, refetch }) => {
    const [removeModalOpen, setRemoveModalOpen] = useState<boolean>(false)
    const [refreshGitHubApp, { loading, error, data }] = useRefreshGitHubApp()
    const createURL = `/site-admin/batch-changes/github-apps/new?baseURL=${encodeURIComponent(baseURL)}`
    const navigate = useNavigate()

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
                <MenuList position={Position.bottomEnd} className={styles.menuList}>
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
