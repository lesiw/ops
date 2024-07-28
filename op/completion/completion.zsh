#compdef op

_ops() {
    local -a ops
    IFS=$'\n'
    ops=($(op -l 2>/dev/null))
    _describe 'ops' ops
}

_arguments \
    '-l[List ops.]' \
    '*:ops:_ops'
