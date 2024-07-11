#compdef ci

_ci_tasks() {
    local -a tasks
    IFS=$'\n'
    tasks=($(ci -l 2>/dev/null))
    _describe 'tasks' tasks
}

_arguments \
    '-l[List tasks.]' \
    '*:task:_ci_tasks'
