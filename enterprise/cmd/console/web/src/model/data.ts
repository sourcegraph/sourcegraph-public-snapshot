export type ConsoleData = ConsoleAnonymousData | ConsoleUserData

export interface ConsoleAnonymousData {
    user: null
}

export interface ConsoleUserData {
    user: UserData
    instances: InstanceData[]
}

export interface UserData {
    email: string
}

export interface InstanceData {
    id: string
    url: string
    ownerEmail: string
    ownerName: string
    viewerIsOwner: boolean
    viewerIsOrganizationMember: boolean
    viewerCanMaybeSignIn: boolean
    status: 'waiting' | 'ready'
}
