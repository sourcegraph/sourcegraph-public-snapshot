#ifdef _WIN32
#include <direct.h>
#define getcwd _getcwd
#define S_ISDIR(mode) (((mode) & S_IFMT) == S_IFDIR)
#else
#include <unistd.h>
#endif

#include <iostream>
#include <fstream>
#include <cctype>
#include <vector>
#include <algorithm>
#include <sys/stat.h>
#include "file.hpp"
#include "context.hpp"
#include "utf8_string.hpp"
#include "sass2scss.h"

#ifdef _WIN32
#include <windows.h>
#endif

#ifndef FS_CASE_SENSITIVE
#ifdef _WIN32
#define FS_CASE_SENSITIVE 0
#else
#define FS_CASE_SENSITIVE 1
#endif
#endif

namespace Sass {
  namespace File {
    using namespace std;

    // return the current directory
    // always with forward slashes
    string get_cwd()
    {
      const size_t wd_len = 1024;
      char wd[wd_len];
      string cwd = getcwd(wd, wd_len);
      #ifdef _WIN32
        //convert backslashes to forward slashes
        replace(cwd.begin(), cwd.end(), '\\', '/');
      #endif
      if (cwd[cwd.length() - 1] != '/') cwd += '/';
      return cwd;
    }

    // test if path exists and is a file
    bool file_exists(const string& path)
    {
      #ifdef _WIN32
        wstring wpath = UTF_8::convert_to_utf16(path);
        DWORD dwAttrib = GetFileAttributesW(wpath.c_str());
        return (dwAttrib != INVALID_FILE_ATTRIBUTES &&
               (!(dwAttrib & FILE_ATTRIBUTE_DIRECTORY)));
      #else
        struct stat st_buf;
        return (stat (path.c_str(), &st_buf) == 0) &&
               (!S_ISDIR (st_buf.st_mode));
      #endif
    }

    // return if given path is absolute
    // works with *nix and windows paths
    bool is_absolute_path(const string& path)
    {
      #ifdef _WIN32
        if (path.length() >= 2 && isalpha(path[0]) && path[1] == ':') return true;
      #endif
      return path[0] == '/';
    }

    // helper function to find the last directory seperator
    inline size_t find_last_folder_separator(const string& path, size_t limit = string::npos)
    {
      size_t pos = string::npos;
      size_t pos_p = path.find_last_of('/', limit);
      #ifdef _WIN32
        size_t pos_w = path.find_last_of('\\', limit);
      #else
        size_t pos_w = string::npos;
      #endif
      if (pos_p != string::npos && pos_w != string::npos) {
        pos = max(pos_p, pos_w);
      }
      else if (pos_p != string::npos) {
        pos = pos_p;
      }
      else {
        pos = pos_w;
      }
      return pos;
    }

    // return only the directory part of path
    string dir_name(const string& path)
    {
      size_t pos = find_last_folder_separator(path);
      if (pos == string::npos) return "";
      else return path.substr(0, pos+1);
    }

    // return only the filename part of path
    string base_name(const string& path)
    {
      size_t pos = find_last_folder_separator(path);
      if (pos == string::npos) return path;
      else return path.substr(pos+1);
    }

    // do a locigal clean up of the path
    // no physical check on the filesystem
    string make_canonical_path (string path)
    {

      // declarations
      size_t pos;

      #ifdef _WIN32
        //convert backslashes to forward slashes
        replace(path.begin(), path.end(), '\\', '/');
      #endif

      pos = 0; // remove all self references inside the path string
      while((pos = path.find("/./", pos)) != string::npos) path.erase(pos, 2);

      pos = 0; // remove all leading and trailing self references
      while(path.length() > 1 && path.substr(0, 2) == "./") path.erase(0, 2);
      while((pos = path.length()) > 1 && path.substr(pos - 2) == "/.") path.erase(pos - 2);

      pos = 0; // collapse multiple delimiters into a single one
      while((pos = path.find("//", pos)) != string::npos) path.erase(pos, 1);

      return path;

    }

    // join two path segments cleanly together
    // but only if right side is not absolute yet
    string join_paths(string l, string r)
    {

      #ifdef _WIN32
        // convert Windows backslashes to URL forward slashes
        replace(l.begin(), l.end(), '\\', '/');
        replace(r.begin(), r.end(), '\\', '/');
      #endif

      if (l.empty()) return r;
      if (r.empty()) return l;

      if (is_absolute_path(r)) return r;
      if (l[l.length()-1] != '/') l += '/';

      while ((r.length() > 3) && ((r.substr(0, 3) == "../") || (r.substr(0, 3)) == "..\\")) {
        r = r.substr(3);
        size_t pos = find_last_folder_separator(l, l.length() - 2);
        l = l.substr(0, pos == string::npos ? pos : pos + 1);
      }

      return l + r;
    }

    // create an absolute path by resolving relative paths with cwd
    string make_absolute_path(const string& path, const string& cwd)
    {
      return make_canonical_path((is_absolute_path(path) ? path : join_paths(cwd, path)));
    }

    // create a path that is relative to the given base directory
    // path and base will first be resolved against cwd to make them absolute
    string resolve_relative_path(const string& uri, const string& base, const string& cwd)
    {

      string absolute_uri = make_absolute_path(uri, cwd);
      string absolute_base = make_absolute_path(base, cwd);

      #ifdef _WIN32
        // absolute link must have a drive letter, and we know that we
        // can only create relative links if both are on the same drive
        if (absolute_base[0] != absolute_uri[0]) return absolute_uri;
      #endif

      string stripped_uri = "";
      string stripped_base = "";

      size_t index = 0;
      size_t minSize = min(absolute_uri.size(), absolute_base.size());
      for (size_t i = 0; i < minSize; ++i) {
        #ifdef FS_CASE_SENSITIVE
          if (absolute_uri[i] != absolute_base[i]) break;
        #else
          // compare the charactes in a case insensitive manner
          // windows fs is only case insensitive in ascii ranges
          if (tolower(absolute_uri[i]) != tolower(absolute_base[i])) break;
        #endif
        if (absolute_uri[i] == '/') index = i + 1;
      }
      for (size_t i = index; i < absolute_uri.size(); ++i) {
        stripped_uri += absolute_uri[i];
      }
      for (size_t i = index; i < absolute_base.size(); ++i) {
        stripped_base += absolute_base[i];
      }

      size_t left = 0;
      size_t directories = 0;
      for (size_t right = 0; right < stripped_base.size(); ++right) {
        if (stripped_base[right] == '/') {
          if (stripped_base.substr(left, 2) != "..") {
            ++directories;
          }
          else if (directories > 1) {
            --directories;
          }
          else {
            directories = 0;
          }
          left = right + 1;
        }
      }

      string result = "";
      for (size_t i = 0; i < directories; ++i) {
        result += "../";
      }
      result += stripped_uri;

      return result;
    }

    // Resolution order for ambiguous imports:
    // (1) filename as given
    // (2) underscore + given
    // (3) underscore + given + extension
    // (4) given + extension
    string resolve_file(const string& filename)
    {
      // supported extensions
      const vector<string> exts = {
        ".scss", ".sass", ".css"
      };
      // split the filename
      string base(dir_name(filename));
      string name(base_name(filename));
      // create full path (maybe relative)
      string path(join_paths(base, name));
      if (file_exists(path)) return path;
      // next test variation with underscore
      path = join_paths(base, "_" + name);
      if (file_exists(path)) return path;
      // next test exts plus underscore
      for(auto ext : exts) {
        path = join_paths(base, "_" + name + ext);
        if (file_exists(path)) return path;
      }
      // next test plain name with exts
      for(auto ext : exts) {
        path = join_paths(base, name + ext);
        if (file_exists(path)) return path;
      }
      // nothing found
      return string("");
    }

    // helper function to resolve a filename
    string find_file(const string& file, const vector<string> paths)
    {
      // search in every include path for a match
      for (size_t i = 0, S = paths.size(); i < S; ++i)
      {
        string path(join_paths(paths[i], file));
        string resolved(resolve_file(path));
        if (resolved != "") return resolved;
      }
      // nothing found
      return string("");
    }

    // inc paths can be directly passed from C code
    string find_file(const string& file, const char* paths[])
    {
      if (paths == 0) return string("");
      vector<string> includes(0);
      // includes.push_back(".");
      const char** it = paths;
      while (it && *it) {
        includes.push_back(*it);
        ++it;
      }
      return find_file(file, includes);
    }

    // try to load the given filename
    // returned memory must be freed
    // will auto convert .sass files
    char* read_file(const string& path)
    {
      #ifdef _WIN32
        BYTE* pBuffer;
        DWORD dwBytes;
        // windows unicode filepaths are encoded in utf16
        wstring wpath = UTF_8::convert_to_utf16(path);
        HANDLE hFile = CreateFileW(wpath.c_str(), GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, 0, NULL);
        if (hFile == INVALID_HANDLE_VALUE) return 0;
        DWORD dwFileLength = GetFileSize(hFile, NULL);
        if (dwFileLength == INVALID_FILE_SIZE) return 0;
        // allocate an extra byte for the null char
        pBuffer = (BYTE*)malloc((dwFileLength+1)*sizeof(BYTE));
        ReadFile(hFile, pBuffer, dwFileLength, &dwBytes, NULL);
        pBuffer[dwFileLength] = '\0';
        CloseHandle(hFile);
        // just convert from unsigned char*
        char* contents = (char*) pBuffer;
      #else
        struct stat st;
        if (stat(path.c_str(), &st) == -1 || S_ISDIR(st.st_mode)) return 0;
        ifstream file(path.c_str(), ios::in | ios::binary | ios::ate);
        char* contents = 0;
        if (file.is_open()) {
          size_t size = file.tellg();
          // allocate an extra byte for the null char
          contents = (char*) malloc((size+1)*sizeof(char));
          file.seekg(0, ios::beg);
          file.read(contents, size);
          contents[size] = '\0';
          file.close();
        }
      #endif
      string extension;
      if (path.length() > 5) {
        extension = path.substr(path.length() - 5, 5);
      }
      for(size_t i=0; i<extension.size();++i)
        extension[i] = tolower(extension[i]);
      if (extension == ".sass" && contents != 0) {
        char * converted = sass2scss(contents, SASS2SCSS_PRETTIFY_1 | SASS2SCSS_KEEP_COMMENT);
        free(contents); // free the indented contents
        return converted; // should be freed by caller
      } else {
        return contents;
      }
    }

  }
}
