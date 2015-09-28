#ifndef SASS_FILE_H
#define SASS_FILE_H

#include <string>
#include <vector>

namespace Sass {
  using namespace std;
  class Context;
  namespace File {

    // return the current directory
    // always with forward slashes
    string get_cwd();

    // test if path exists and is a file
    bool file_exists(const string& file);

    // return if given path is absolute
    // works with *nix and windows paths
    bool is_absolute_path(const string& path);

    // return only the directory part of path
    string dir_name(const string& path);

    // return only the filename part of path
    string base_name(const string&);

    // do a locigal clean up of the path
    // no physical check on the filesystem
    string make_canonical_path (string path);

    // join two path segments cleanly together
    // but only if right side is not absolute yet
    string join_paths(string root, string name);

    // create an absolute path by resolving relative paths with cwd
    string make_absolute_path(const string& path, const string& cwd = ".");

    // create a path that is relative to the given base directory
    // path and base will first be resolved against cwd to make them absolute
    string resolve_relative_path(const string& path, const string& base, const string& cwd = ".");

    // try to find/resolve the filename
    string resolve_file(const string& file);

    // helper function to resolve a filename
    string find_file(const string& file, const vector<string> paths);
    // inc paths can be directly passed from C code
    string find_file(const string& file, const char** paths);

    // try to load the given filename
    // returned memory must be freed
    // will auto convert .sass files
    char* read_file(const string& file);

  }
}

#endif
