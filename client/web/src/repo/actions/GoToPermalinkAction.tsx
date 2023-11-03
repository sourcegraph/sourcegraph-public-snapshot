import React, { useEffect, useMemo, useState } from 'react'

import { mdiLink, mdiChevronDown, mdiContentCopy, mdiCheckBold } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import { useLocation, useNavigate } from 'react-router-dom'
import { fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'
import {
    Position,
    Icon,
    Link,
    Tooltip,
    Button,
    Text,
    Menu,
    ButtonGroup,
    MenuButton,
    MenuList,
    MenuItem,
    screenReaderAnnounce,
} from '@sourcegraph/wildcard'

import { replaceRevisionInURL } from '../../util/url'
import {
    RepoHeaderActionButtonLink,
    RepoHeaderActionMenuLink,
    RepoHeaderActionDropdownToggle,
} from '../components/RepoHeaderActions'
import type { RepoHeaderContext } from '../RepoHeader'

import styles from './actions.module.scss'

interface GoToPermalinkActionProps extends RepoHeaderContext, TelemetryProps {
    /**
     * The current (possibly undefined or non-full-SHA) Git revision.
     */
    revision?: string

    /**
     * The commit SHA for the revision in the current location (URL).
     */
    commitID?: string
}

/**
 * A repository header action that replaces the revision in the URL with the canonical 40-character
 * Git commit SHA.
 */
export const GoToPermalinkAction: React.FunctionComponent<GoToPermalinkActionProps> = props => {
    const { revision, commitID, actionType, repoName, telemetryService } = props

    const navigate = useNavigate()
    const location = useLocation()
    const fullURL = location.pathname + location.search + location.hash
    const permalinkURL = useMemo(() => replaceRevisionInURL(fullURL, commitID || ''), [fullURL, commitID])
    const [copied, setCopied] = useState<boolean>(false)

    useEffect(() => {
        // Trigger the user presses 'y'.
        const subscription = fromEvent<KeyboardEvent>(window, 'keydown')
            .pipe(
                filter(
                    event =>
                        // 'y' shortcut (if no input element is focused)
                        event.key === 'y' && !!document.activeElement && !isInputElement(document.activeElement)
                )
            )
            .subscribe(event => {
                event.preventDefault()

                // Replace the revision in the current URL with the new one and push to history.
                navigate(permalinkURL)
            })

        return () => subscription.unsubscribe()
    }, [navigate, permalinkURL])

    const onClick = (): void => {
        telemetryService.log('PermalinkClicked', { repoName, commitID })
    }

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink as={Link} file={true} to={permalinkURL} onSelect={onClick}>
                <Icon aria-hidden={true} svgPath={mdiLink} />
                <span>Permalink (with full Git commit SHA)</span>
            </RepoHeaderActionMenuLink>
        )
    }

    const copyPermalink = (event: React.MouseEvent<HTMLButtonElement>): void => {
        event.preventDefault()
        telemetryService.log('CopyFilePath')
        copy(permalinkURL)
        setCopied(true)
        screenReaderAnnounce('Path copied to clipboard')

        setTimeout(() => {
            setCopied(false)
        }, 1000)
    }

    const copyLinkLabel = copied ? 'Copied!' : 'Copy Link'
    const copyLinkIcon = copied ? mdiCheckBold : mdiContentCopy
    const isRevisionTheSameAsCommitID = revision === commitID
    console.log({ revision, commitID })

    return (
        <Tooltip content="Permalink (with full Git commit SHA)">
            {/* <RepoHeaderActionDropdownToggle aria-label="Permalink"> */}
            <Menu>
                <ButtonGroup>
                    <Button className={classNames('border', styles.permalinkBtn)} onClick={copyPermalink}>
                        <Icon
                            aria-hidden={true}
                            svgPath={copyLinkIcon}
                            class={classNames(styles.copyIcon, {
                                [styles.checkIcon]: copied,
                            })}
                        />
                        <Text className={classNames(styles.repoActionLabel, 'text-muted ml-0')}>{copyLinkLabel}</Text>
                    </Button>
                    {!isRevisionTheSameAsCommitID && (
                        <MenuButton variant="secondary" className={styles.chevronBtn}>
                            <Icon
                                className={styles.chevronBtnIcon}
                                svgPath={mdiChevronDown}
                                inline={false}
                                aria-hidden={true}
                            />
                            <VisuallyHidden>Actions</VisuallyHidden>
                        </MenuButton>
                    )}
                    {!isRevisionTheSameAsCommitID && (
                        <MenuList position={Position.bottomEnd}>
                            <MenuItem onSelect={copyPermalink} className={styles.dropdownItem}>
                                <Text>Copy permalink</Text>
                            </MenuItem>
                        </MenuList>
                    )}
                </ButtonGroup>
            </Menu>
            {/* </RepoHeaderActionDropdownToggle> */}
            {/* <RepoHeaderActionButtonLink aria-label="Permalink" file={false} to={permalinkURL} onSelect={onClick}>
                <Icon aria-hidden={true} svgPath={mdiLink} />
            </RepoHeaderActionButtonLink> */}
        </Tooltip>
    )
}
