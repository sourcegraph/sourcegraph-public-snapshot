export const langs: [string, string][] = [
	["golang", "Go"],
	["java", "Java"],
	["javaandroid", "Java (Android)"],
	["python", "Python"],
	["javascript", "JavaScript"],
	["typescript", "TypeScript"],
	["c", "C"],
	["cpp", "C++"],
	["csharp", "C#"],
	["php", "PHP"],
	["scala", "Scala"],
	["swift", "Swift"],
	["objectivec", "Objective-C"],
	["rust", "Rust"],
	["ruby", "Ruby"],
	["other", "Other"],
];

export function langIDs(): string[] {
	return langs.map(([id]) => id);
}

export function langNames(): string[] {
	return langs.map(([id, name]) => name);
}

export function langName(lang: string): string {
	const found = langs.find(([id, name]) => id === lang);
	if (!found) { throw new Error(`language name not found: ${JSON.stringify(lang)}`); }
	return found[1];
}
