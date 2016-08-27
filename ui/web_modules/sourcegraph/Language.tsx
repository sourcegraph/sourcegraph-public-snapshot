export const allLangs: string[] = ["golang", "java", "python", "javascript", "other"];

export const supportedLangs: string[] = ["golang", "java", "other"];
export const langIsSupported = (lang: string) => supportedLangs.includes(lang);

const names: {[key: string]: string} = {
	golang: "Go",
	java: "Java",
	python: "Python",
	javascript: "JavaScript",
	other: "Other",
};
export const langName = (lang: string) => names[lang];
