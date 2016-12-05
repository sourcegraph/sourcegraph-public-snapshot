declare module "editorconfig" {
    export interface FileInfo {
        indent_style?: string;
        indent_size?: number;
        tab_width?: number;
        end_of_line?: string;
        charset?: string;
        trim_trailing_whitespace?: boolean;
        insert_final_newline?: boolean;
        root?: string;
    }

    export interface ParseOptions {
        /* config file name. default: .editorconfig */
        config: string;
        version: any; // string or Version
    }

    export function parseFromFiles(filepath: string, files: any[], options?: ParseOptions): Promise<FileInfo>;

    export function parse(filepath: string, options?: ParseOptions): Promise<FileInfo>;
}
