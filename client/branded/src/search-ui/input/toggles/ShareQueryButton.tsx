import React, { useRef, useState } from 'react'

import { mdiClose, mdiContentCopy, mdiExportVariant } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import {
    Button,
    Icon,
    Input,
    Label,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Tooltip,
    Position,
    H4,
    H2,
    Select,
    Form,
    Link,
} from '@sourcegraph/wildcard'

import shareStyles from './ShareQueryButton.module.scss'
import styles from './Toggles.module.scss'

interface Props {
    authenticatedUser: AuthenticatedUser | null
    onSaveSearch: (destination: SaveQueryDestination, title: string, fullQuery: string) => Promise<string>
    className?: string
    fullQuery: string
}

/**
 * A toggle displayed in the QueryInput.
 */
export const ShareQueryButton: React.FunctionComponent<Props> = ({
    authenticatedUser,
    onSaveSearch,
    className,
    fullQuery,
}) => {
    const tooltipValue = 'Save and share this query'
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)

    return (
        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
            <Tooltip content={tooltipValue} placement="bottom">
                <PopoverTrigger
                    as={Button}
                    className={classNames('a11y-ignore', styles.toggle, shareStyles.button, className)}
                    variant="icon"
                >
                    <Icon aria-label={tooltipValue} svgPath={mdiExportVariant} />
                </PopoverTrigger>
            </Tooltip>

            <ShareQueryButtonMenu
                authenticatedUser={authenticatedUser}
                fullQuery={fullQuery}
                onSaveSearch={onSaveSearch}
                closeMenu={() => setIsPopoverOpen(false)}
            />
        </Popover>
    )
}

export interface SaveQueryDestination {
    __typename: string
    id: string
    displayName?: string | null
    name?: string
}

const ShareQueryButtonMenu: React.FunctionComponent<Props & { closeMenu: () => void }> = ({
    authenticatedUser,
    fullQuery,
    onSaveSearch,
    closeMenu,
}) => {
    let [destination, setDestination] = useState<SaveQueryDestination>(authenticatedUser!)
    let [title, setTitle] = useState<string>('')
    let [shortUrl, setShortUrl] = useState<string>('')
    let [error, setError] = useState<string>('')

    const shortUrlCopyRef = useRef<HTMLInputElement | null>(null)

    const reset = () => {
        setDestination(authenticatedUser!)
        setTitle('')
        setShortUrl('')
        setError('')
    }

    let savedSearchesLink = (
        <Link
            to={
                destination === authenticatedUser
                    ? `${location.origin}/users/${authenticatedUser!.username}/searches`
                    : `${location.origin}/organizations/${destination.name}/searches`
            }
        >
            {destination === authenticatedUser
                ? 'your'
                : `${
                      destination.displayName && destination.displayName.length > 0
                          ? destination.displayName
                          : destination.name
                  }'s`}{' '}
            saved searches
        </Link>
    )

    return (
        <PopoverContent
            aria-labelledby="smart-search-popover-header"
            position={Position.bottomEnd}
            className={shareStyles.popoverWindow}
            onOpenChange={() => {
                reset()
                closeMenu()
            }}
        >
            <div className={classNames(shareStyles.popoverWindowTop, 'd-flex align-items-center px-3 py-2')}>
                <H4 as={H2} id="smart-search-popover-header" className="m-0 flex-1">
                    Save and Share
                </H4>
                <Button
                    onClick={() => {
                        reset()
                        closeMenu()
                    }}
                    variant="icon"
                    aria-label="Close"
                >
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            {!shortUrl && !error && (
                <Form
                    onSubmit={event => {
                        onSaveSearch(destination, title, fullQuery)
                            .then(url => {
                                setShortUrl(url)
                            })
                            .catch(err => {
                                setError(err.toString())
                            })
                        event.stopPropagation()
                        event.preventDefault()
                    }}
                    className="d-flex flex-column px-3 py-2"
                >
                    <Label className={shareStyles.label} id="save-destination-label">
                        <span className="d-flex flex-column mb-2">
                            <span className={shareStyles.labelHeader}>Destination</span>
                            <span className={shareStyles.labelDescription}>
                                Where would you like to save this search? This could be your own profile or an
                                organization you're a member of.
                            </span>
                        </span>
                        <Select
                            onChange={event => {
                                setDestination(
                                    event.target.selectedIndex === 0
                                        ? authenticatedUser!
                                        : authenticatedUser!.organizations.nodes[event.target.selectedIndex - 1]
                                )
                            }}
                            aria-labelledby="save-destination-label"
                            isCustomStyle={true}
                            className="mb-0"
                        >
                            {[
                                {
                                    id: authenticatedUser!.id,
                                    name:
                                        authenticatedUser!.displayName && authenticatedUser!.displayName.length > 0
                                            ? authenticatedUser!.displayName
                                            : authenticatedUser!.username,
                                },
                                ...authenticatedUser!.organizations.nodes.map(({ id, displayName, name }) => ({
                                    id,
                                    name: displayName && displayName.length > 0 ? displayName : name,
                                })),
                            ].map(org => (
                                <option key={org.id} value={org.id} label={org.name} />
                            ))}
                        </Select>
                    </Label>
                    <Label className={shareStyles.label}>
                        <span className="d-flex flex-column mb-2">
                            <span className={shareStyles.labelHeader}>Title</span>
                            <span className={shareStyles.labelDescription}>
                                Give this search a snazzy title so you can easily find it in {savedSearchesLink}.
                            </span>
                        </span>
                        <Input
                            required
                            onChange={event => {
                                setTitle(event.target.value)
                            }}
                        />
                    </Label>
                    <Button type="submit" variant="primary">
                        Save
                    </Button>
                </Form>
            )}
            {shortUrl && (
                <div className="d-flex flex-column px-3 py-2">
                    <Label className={shareStyles.label}>
                        <span className="d-flex flex-column mb-2">
                            <span className={shareStyles.labelHeader}>Short URL</span>
                            <span className={shareStyles.labelDescription}>
                                Use this short link to access this search anytime. You can also modify or delete this
                                saved search in {savedSearchesLink}.
                            </span>
                        </span>
                        <div className="input-group">
                            <Input readOnly={true} value={shortUrl} type="url" ref={shortUrlCopyRef} />
                            <div className="input-group-append">
                                <Button
                                    variant="secondary"
                                    aria-label="Copy"
                                    onClick={() => {
                                        shortUrlCopyRef.current!.select()
                                        copy(shortUrl)
                                    }}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiContentCopy} />
                                </Button>
                            </div>
                        </div>
                    </Label>
                </div>
            )}
            {error && <span className="d-flex flex-column px-3 py-2">An error occurred: {error}</span>}
        </PopoverContent>
    )
}
