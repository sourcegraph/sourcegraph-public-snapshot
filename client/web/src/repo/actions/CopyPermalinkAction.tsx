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
    Button,
    Menu,
    ButtonGroup,
    MenuButton,
    MenuList,
    MenuItem,
    screenReaderAnnounce,
} from '@sourcegraph/wildcard'

import { replaceRevisionInURL } from '../../util/url'
import { RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'
import { RepoActionInfo } from '../RepoActionInfo'
import type { RepoHeaderContext } from '../RepoHeader'

import styles from './actions.module.scss'

interface CopyPermalinkActionProps extends RepoHeaderContext, TelemetryProps {
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
export const CopyPermalinkAction: React.FunctionComponent<CopyPermalinkActionProps> = props => {
    const { revision, commitID, actionType, repoName, telemetryService } = props

    const navigate = useNavigate()
    const location = useLocation()
    const fullURL = location.pathname + location.search + location.hash
    const permalinkURL = useMemo(() => replaceRevisionInURL(fullURL, commitID || ''), [fullURL, commitID])
    const linkURL = useMemo(() => replaceRevisionInURL(fullURL, revision || ''), [fullURL, revision])
    const [copiedPermalink, setCopiedPermalink] = useState<boolean>(false)
    const [copiedLink, setCopiedLink] = useState<boolean>(false)

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

    const copyPermalink = (): void => {
        telemetryService.log('CopyPermalink')
        copy(permalinkURL)
        setCopiedPermalink(true)
        screenReaderAnnounce('Permalink copied to clipboard')

        setTimeout(() => setCopiedPermalink(false), 1000)
    }

    const copyLink = (): void => {
        telemetryService.log('CopyLink')
        copy(linkURL)
        setCopiedLink(true)
        screenReaderAnnounce('Link copied to clipboard')

        setTimeout(() => setCopiedLink(false), 1000)
    }

    const isRevisionTheSameAsCommitID = revision === commitID
    const copyLinkLabel = copiedLink ? 'Copied!' : isRevisionTheSameAsCommitID ? 'Copy Link' : 'Links'
    const copyLinkIcon = copiedLink ? mdiCheckBold : mdiContentCopy

    return (
        <Menu>
            <ButtonGroup>
                <Button className={classNames('border', styles.permalinkBtn, 'pt-0 pb-0')} onClick={copyLink}>
                    <RepoActionInfo
                        displayName={copyLinkLabel}
                        icon={
                            <Icon
                                svgPath={copyLinkIcon}
                                aria-hidden={true}
                                className={classNames({
                                    [styles.checkedIcon]: copiedLink,
                                    [styles.repoActionIcon]: !copiedLink,
                                })}
                            />
                        }
                    />
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
                        <MenuItem
                            onSelect={copyPermalink}
                            className={classNames(styles.dropdownItem, 'justify-content-start')}
                        >
                            <RepoActionInfo
                                displayName={copiedPermalink ? 'Copied' : 'Copy permalink'}
                                icon={
                                    <Icon
                                        aria-hidden={true}
                                        svgPath={copiedPermalink ? mdiCheckBold : mdiContentCopy}
                                        className={classNames({
                                            [styles.checkedIcon]: copiedLink,
                                            [styles.repoActionIcon]: !copiedLink,
                                        })}
                                    />
                                }
                                className={styles.permalinkText}
                            />
                        </MenuItem>
                    </MenuList>
                )}
            </ButtonGroup>
        </Menu>
    )
}
