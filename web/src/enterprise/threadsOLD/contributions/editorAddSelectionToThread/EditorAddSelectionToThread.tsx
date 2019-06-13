import * as H from 'history'
import PlusBoxIcon from 'mdi-react/PlusBoxIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { LineOrPositionOrRange, RepoFile } from '../../../../../../shared/src/util/url'
import { addTargetToThread } from '../../../../discussions/backend'
import { ThreadDropdownMenu } from './ThreadDropdownMenu'

interface Props extends RepoFile, ExtensionsControllerNotificationProps {
    /** The currently selected position. */
    selectedPosition: LineOrPositionOrRange

    /** The position of the tooltip (assigned to `style`) */
    overlayPosition?: { left: number; top: number }

    location: H.Location
}

/**
 * A gutter decoration in the editor that shows a menu for discussing or attaching the current
 * selection.
 */
export const EditorAddSelectionToThread: React.FunctionComponent<Props> = ({
    repoName,
    filePath,
    rev,
    commitID,
    selectedPosition,
    overlayPosition,
    location,
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onAddToExistingThreadClick = useCallback(
        async (thread: Pick<GQL.IDiscussionThread, 'id'>) => {
            try {
                await addTargetToThread({
                    threadID: thread.id,
                    target: {
                        repo: {
                            repositoryName: repoName,
                            branch: rev,
                            path: filePath,
                            revision: commitID,
                            selection:
                                selectedPosition.line !== undefined
                                    ? {
                                          startLine: selectedPosition.line,
                                          startCharacter:
                                              selectedPosition.character !== undefined ? selectedPosition.character : 0,
                                          endLine:
                                              selectedPosition.endLine !== undefined
                                                  ? selectedPosition.endLine
                                                  : selectedPosition.line,
                                          endCharacter:
                                              selectedPosition.endCharacter !== undefined
                                                  ? selectedPosition.endCharacter
                                                  : selectedPosition.character || 0,
                                      }
                                    : null,
                        },
                    },
                }).toPromise()
                extensionsController.services.notifications.showMessages.next({
                    message: `Attached!`,
                    type: NotificationType.Success,
                })
                // HACK TODO!(sqs): auto-close this
                setTimeout(() => {
                    const notification = document.querySelector('.sourcegraph-notification-item.alert-success')
                    if (notification) {
                        const closeButton = notification.querySelector<HTMLButtonElement>('button.close')
                        if (closeButton) {
                            closeButton.click()
                        }
                    }
                }, 1000)
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error attaching to thread: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [
            commitID,
            extensionsController.services.notifications.showMessages,
            filePath,
            repoName,
            rev,
            selectedPosition.character,
            selectedPosition.endCharacter,
            selectedPosition.endLine,
            selectedPosition.line,
        ]
    )

    return (
        <div
            className="discussions-gutter-overlay"
            // tslint:disable-next-line:jsx-ban-props needed for dynamic styling
            style={
                overlayPosition
                    ? {
                          position: 'absolute',
                          opacity: 1,
                          visibility: 'visible',
                          left: overlayPosition.left + 'px',
                          top: overlayPosition.top - 3 + 'px', // TODO!(sqs): hack 4px
                      }
                    : {
                          opacity: 0,
                          visibility: 'hidden',
                      }
            }
        >
            <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
                <DropdownToggle
                    className="btn btn-sm btn-link btn-icon"
                    data-tooltip={isOpen ? undefined : 'Comment or attach...'}
                >
                    <PlusBoxIcon className="icon-inline" />
                </DropdownToggle>
                <ThreadDropdownMenu
                    onAddToExistingThreadClick={onAddToExistingThreadClick}
                    right={true}
                    location={location}
                />
            </ButtonDropdown>
        </div>
    )
}
