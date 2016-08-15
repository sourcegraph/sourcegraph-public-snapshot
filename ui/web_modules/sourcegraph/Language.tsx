export const allLangs: string[] = ["golang", "java", "python", "javascript"];

export const supportedLangs: string[] = ["golang", "java"];
export const langIsSupported = (lang: string) => supportedLangs.includes(lang);

const names: {[key: string]: string} = {
	golang: "Go",
	java: "Java",
	python: "Python",
	javascript: "JavaScript",
};
export const langName = (lang: string) => names[lang];
