import { readFileSync } from 'fs'

export interface User {
	name: string
	email: string
	accessToken: string
	accessibleCodebaseIDs: string[]
}

const BEARER_PREFIX = 'Bearer '

export function authenticate(
	authorizationHeader: string | undefined,
	urlQueryToken: string | null,
	users: User[]
): User | null {
	const headerToken = authorizationHeader?.startsWith(BEARER_PREFIX)
		? authorizationHeader.slice(BEARER_PREFIX.length).trim()
		: null
	const tokenToValidate = headerToken ?? urlQueryToken // prefer header
	if (!tokenToValidate) {
		return null // no token
	}
	const user = users.find(user => user.accessToken === tokenToValidate)
	return user ?? null
}

export function getUsers(usersPath: string): User[] {
	return JSON.parse(readFileSync(usersPath).toString()) as User[]
}
