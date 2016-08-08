export interface IConfigurationFile {
    extends?: string | string[];
    linterOptions?: {
        typeCheck?: boolean;
    };
    rulesDirectory?: string | string[];
    rules?: any;
}
export declare const CONFIG_FILENAME: string;
export declare const DEFAULT_CONFIG: {
    "rules": {
        "class-name": boolean;
        "comment-format": (boolean | string)[];
        "indent": (boolean | string)[];
        "no-duplicate-variable": boolean;
        "no-eval": boolean;
        "no-internal-module": boolean;
        "no-trailing-whitespace": boolean;
        "no-unsafe-finally": boolean;
        "no-var-keyword": boolean;
        "one-line": (boolean | string)[];
        "quotemark": (boolean | string)[];
        "semicolon": (boolean | string)[];
        "triple-equals": (boolean | string)[];
        "typedef-whitespace": (boolean | {
            "call-signature": string;
            "index-signature": string;
            "parameter": string;
            "property-declaration": string;
            "variable-declaration": string;
        })[];
        "variable-name": (boolean | string)[];
        "whitespace": (boolean | string)[];
    };
};
export declare function findConfiguration(configFile: string, inputFilePath: string): IConfigurationFile;
export declare function findConfigurationPath(suppliedConfigFilePath: string, inputFilePath: string): string;
export declare function loadConfigurationFromPath(configFilePath: string): IConfigurationFile;
export declare function extendConfigurationFile(config: IConfigurationFile, baseConfig: IConfigurationFile): IConfigurationFile;
export declare function getRelativePath(directory: string, relativeTo?: string): string;
export declare function getRulesDirectories(directories: string | string[], relativeTo?: string): string[];
