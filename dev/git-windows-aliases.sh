#!/bin/bash

# Adds the following aliases: 
# git add-symlink SRC DST
# git rm-symlink LNK
# git rm-symlinks
# git checkout-symlinks
# See http://stackoverflow.com/questions/5917249/git-symlinks-in-windows/16754068#16754068

git config --global alias.add-symlink '!__git_add_symlink(){
    argv=($@)
    argc=${#argv[@]}

    # Look for options
    options=(" -h")
    o_help="false"
    case "${argv[@]}" in *" -h"*) o_help="true" ;; esac
    if [ "$o_help" == "true" -o "$argc" -lt "2" ]; then
        echo "\
Usage: git add-symlink <src> <dst>

* <src> is a RELATIVE PATH, respective to <dst>.
* <dst> is a RELATIVE PATH, respective to the repository'\''s root dir.
* Command must be run from the repository'\''s root dir."
        return 0
    fi

    src_arg=${argv[0]}
    dst_arg=${argv[1]}

    dst=$(echo "$dst_arg")/../$(echo "$src_arg")
    if [ ! -e "$dst" ]; then
        echo "ERROR: Target $dst does not exist; not creating invalid symlink."
        return 1
    fi

    hash=$(echo -n "$src_arg" | git hash-object -w --stdin)
    git update-index --add --cacheinfo 120000 "$hash" "$dst_arg"
    git checkout -- "$dst_arg"

}; __git_add_symlink "$@"'

git config --global alias.rm-symlink '!__git_rm_symlink(){ 
    git checkout -- "$1"
    link=$(echo "$1")
    POS=$'\''/'\''
    DOS=$'\''\\\\'\''
    doslink=${link//$POS/$DOS}
    dest=$(cygpath -aw $(dirname "$link"))/$(cat "$link")
    dosdest=${dest//$POS/$DOS}
    if [ -f "$dest" ]; then
        rm -f "$link"
        cmd //C mklink "$doslink" "$dosdest"
    elif [ -d "$dest" ]; then
        rm -f "$link"
        cmd //C mklink //J "$doslink" "$dosdest"
    else
        echo "ERROR: Something went wrong when processing $1 . . ."
        echo "       $dest may not actually exist as a valid target."
    fi
}; __git_rm_symlink "$1"'

git config --global alias.rm-symlinks '!__git_rm_symlinks(){
    for symlink in $(git ls-files -s | egrep "^120000" | cut -f2); do
        git rm-symlink "$symlink"
        git update-index --assume-unchanged "$symlink"
    done
}; __git_rm_symlinks'

git config --global alias.checkout-symlinks '!__git_checkout_symlinks(){
    POS=$'\''/'\''
    DOS=$'\''\\\\'\''
    for symlink in $(git ls-files -s | egrep "^120000" | cut -f2); do
        git update-index --no-assume-unchanged "$symlink"
        dossymlink=${symlink//$POS/$DOS}
        cmd //C rmdir //Q "$dossymlink" 2>/dev/null
        git  checkout -- "$symlink"
        echo "Restored git symlink $symlink <<===>> $(cat $symlink)"
    done
}; __git_checkout_symlinks'
