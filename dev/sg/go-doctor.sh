#!/bin/bash
min_version="1.17.0"
IFS='.  ' read -r -a min_version_arr <<< $min_version
version=$(go env GOVERSION | grep -Po "[0-9.]+")
IFS='.  ' read -r -a curr_version_arr <<< "$version"


for index in "${!min_version_arr[@]}"
do
    # length of minimum version is longer than current version, check if this index contains all zeros
    if [[ $index -ge ${#curr_version_arr[@]} ]] 
    then
        str=${min_version_arr[index]}

        for((i=0;i<${#str};i++))
        do
            if [ "${str:$i:1}" -ne "0" ]
            then
                echo "go version is less than required version" 1>&2
                exit 1
                break 2
            fi
        done
    fi
    
    min_str=${min_version_arr[index]}
    cur_str=${curr_version_arr[index]}
    
    if [[ ${#min_str} -gt ${#cur_str} ]]
    then
        echo "go version is less than required version" 1>&2
        exit 1
        break 2;
    elif [[ ${#min_str} -eq ${#cur_str} ]]
    then
            for((i=0;i<${#min_str};i++))
            do
                if [ "${min_str:$i:1}" -gt "${cur_str:$i:1}" ]
                then
                  echo "go version is less than required version" 1>&2
                  exit 1
                  break 2
                fi
           done
    fi
done