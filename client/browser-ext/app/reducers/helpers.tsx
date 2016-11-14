export function keyFor(repo: string, rev?: string, path?: string, query?: string): string {
	return `${repo || null}@${rev || null}@${path || null}@${query || null}`;
}
