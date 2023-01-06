import { mdiChat, mdiChevronDown, mdiChevronUp, mdiEmail } from '@mdi/js'
import { Button, Icon } from '@sourcegraph/wildcard'
import { FunctionComponent, PropsWithChildren, useCallback, useState } from 'react'

type Props = {
    handle: String
    email: String
    reasons: OwnershipReasonDetails[]
}

type OwnershipReasonDetails = {
    title: String
    description: String
}

export const FileOwnershipReasons: FunctionComponent<PropsWithChildren<Props>> = ({
    handle,
    email,
    reasons,
}) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )
    return (
        <tr>
            <td>
                <div className="d-flex">

                    <Button
                        variant="icon"
                        className="d-none d-sm-block mx-1"
                        aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                        onClick={toggleIsExpanded}
                    >
                        <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
                    </Button>
                    <Button variant="icon" className="mr-2">
                        <Icon svgPath={mdiEmail} aria-label="email" />
                    </Button>
                    <Button variant="icon">
                        <Icon svgPath={mdiChat} aria-label="chat" />
                    </Button>
                </div>
            </td>
            <td>{handle}</td>
            <td>{email}</td>
            <td>{reasons.map(r => r.title).join(', ')}</td>
        </tr>
    )
}
