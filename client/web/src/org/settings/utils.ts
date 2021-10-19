import { AuthenticatedUser } from 'src/auth'

export const showOrganizationsCode = (user: AuthenticatedUser | null): boolean =>
    user?.tags?.includes('OrgsCode') || false
