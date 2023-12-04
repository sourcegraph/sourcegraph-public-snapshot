import type { FunctionComponent } from 'react'

import { mdiChevronDown } from '@mdi/js'
import VisuallyHidden from '@reach/visually-hidden'

import {
    Button,
    ButtonGroup,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuLink,
    MenuList,
    Position,
    Text,
} from '@sourcegraph/wildcard'

import styles from './CreatePolicyButtons.module.scss'

interface CreatePolicyButtonsProps {
    repo?: { id: string; name: string }
}

export const CreatePolicyButtons: FunctionComponent<CreatePolicyButtonsProps> = ({ repo }) => (
    <Menu>
        <ButtonGroup>
            <Button to="./new?type=head" variant="primary" as={Link}>
                Create new {!repo && 'global'} policy
            </Button>
            <MenuButton variant="primary" className={styles.dropdownButton}>
                <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                <VisuallyHidden>Actions</VisuallyHidden>
            </MenuButton>
        </ButtonGroup>
        <MenuList position={Position.bottomEnd} className={styles.dropdownList}>
            <MenuLink as={Link} className={styles.dropdownItem} to="./new?type=head">
                <>
                    <Text weight="medium" className="mb-2">
                        Create new {!repo && 'global'} policy for HEAD
                    </Text>
                    <Text className="mb-0 text-muted">
                        Match the tip of the default branch{' '}
                        {repo ? 'within this repository' : 'across multiple repositories'}
                    </Text>
                </>
            </MenuLink>
            <MenuLink as={Link} className={styles.dropdownItem} to="./new?type=branch">
                <Text weight="medium" className="mb-2">
                    Create new {!repo && 'global'} branch policy
                </Text>
                <Text className="mb-0 text-muted">
                    Match multiple branches {repo ? 'within this repository' : 'across multiple repositories'}
                </Text>
            </MenuLink>
            <MenuLink as={Link} className={styles.dropdownItem} to="./new?type=tag">
                <Text weight="medium" className="mb-2">
                    Create new {!repo && 'global'} tag policy
                </Text>
                <Text className="mb-0 text-muted">
                    Match multiple tags {repo ? 'within this repository' : 'across multiple repositories'}
                </Text>
            </MenuLink>
        </MenuList>
    </Menu>
)
