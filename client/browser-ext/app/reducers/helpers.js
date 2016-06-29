export const keyFor = (repo, rev, path, query) => `${repo || null}@${rev || null}@${path || null}@${query}`;
