const domains = ["sourcegraph.com/", "github.com/"];

export default function stripDomain(path: string): string {
	for (let domain of domains) {
		if (path.indexOf(domain) === 0) {
			return path.substring(domain.length);
		}
	}
	return path;
}
