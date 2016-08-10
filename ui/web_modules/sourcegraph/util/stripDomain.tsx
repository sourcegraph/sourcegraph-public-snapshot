// tslint:disable: typedef ordered-imports curly

const domains = ["sourcegraph.com/", "github.com/"];

export function stripDomain(path: string): string {
	for (let domain of domains) {
		if (path.indexOf(domain) === 0) {
			return path.substring(domain.length);
		}
	}
	return path;
}
