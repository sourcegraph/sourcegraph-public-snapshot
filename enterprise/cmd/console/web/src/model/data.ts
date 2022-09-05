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
    title: string
    url: string
    ownerEmail: string
    ownerName: string
}
