// isExternalLink returns true if given URL is considered as external link.
export function isExternalLink(url: string): bool {
	return (/^https?:\/\/(nodejs\.org|developer\.mozilla\.org)/).test(url);
}
