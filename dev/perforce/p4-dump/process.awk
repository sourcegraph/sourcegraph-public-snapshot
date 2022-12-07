BEGIN {
    delete non_group_users[0]
    delete groups[0]
}

NF == 1 {
    non_group_users[length(non_group_users) + 1]=$1
}

NF == 2 {
    groups[$2][length(groups[$2]) + 1]=$1;
}

END {
    asorti(groups, sorted_groups)

    for (i in sorted_groups) {
        group=sorted_groups[i]

        printf "%s:", group

        asort(groups[group], sorted_users)
        for (j in sorted_users) {

            printf " %s", sorted_users[j]
        }
       printf "\n"

    }

    if (length(non_group_users) >= 1) {
        printf "---\n"

        printf "The following users are not members of any group:"

        asort(non_group_users, sorted_non_group_users)

        for (i in sorted_non_group_users) {
            printf " %s", sorted_non_group_users[i]
        }

        printf "\n"
    }

    printf "\n"
}
