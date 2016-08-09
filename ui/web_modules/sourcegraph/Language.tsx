export const allLangs: Array<string> = ["golang", "java", "python", "javascript"];

export const supportedLangs: Array<string> = ["golang", "java"];
export const langIsSupported = (lang: string) => supportedLangs.includes(lang);

const names: {[key: string]: string} = {
	golang: "Go",
	java: "Java",
	python: "Python",
	javascript: "JavaScript",
};
export const langName = (lang: string) => names[lang];
