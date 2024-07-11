#!/usr/bin/env bash

__ci_completion () {
    case "${COMP_WORDS[COMP_CWORD]}" in
        -*) suggestions="-l"
            ;;
        *)
            suggestions="$(ci -l)"
            ;;
    esac
    [ -z "$suggestions" ] && return 0
    COMPREPLY=()
    while read -r suggestion
    do
        COMPREPLY+=("$suggestion")
    done < <(compgen -W "$suggestions" -- "${COMP_WORDS[COMP_CWORD]}")
}

complete -F __ci_completion ci
