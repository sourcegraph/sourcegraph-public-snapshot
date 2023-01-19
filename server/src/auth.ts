import { readFileSync } from 'fs'

interface User {
	name: string
	email: string
	accessToken: string
	accessibleCodebaseIDs: string[]
}

const BEARER_PREFIX = 'Bearer '

export function authenticate(authorizationHeader: string | undefined, users: User[]): User | null {
	if (!authorizationHeader) {
		return null
	}
	if (!authorizationHeader.startsWith(BEARER_PREFIX)) {
		return null
	}
	const token = authorizationHeader.slice(BEARER_PREFIX.length).trim()
	const user = users.find(user => user.accessToken === token)
	return user ?? null
}

export function getUsers(usersPath: string): User[] {
	return JSON.parse(readFileSync(usersPath).toString()) as User[]
}
