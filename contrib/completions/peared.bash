# bash completion for peared

_peared_compopt() {
        if type compopt >/dev/null 2>&1; then
                compopt "$@"
        fi
}

_peared_adapter_candidates() {
        if ! command -v peared >/dev/null 2>&1; then
                return
        fi

        peared adapters list 2>/dev/null | awk '{print $1}'
}

_peared_complete_files() {
        local cur_word="$1"
        _peared_compopt -o filenames 2>/dev/null
        COMPREPLY=( $(compgen -f -- "$cur_word") )
}

_peared_complete_dirs() {
        local cur_word="$1"
        _peared_compopt -o filenames 2>/dev/null
        COMPREPLY=( $(compgen -d -- "$cur_word") )
}

_peared_complete_adapters() {
        local cur_word="$1"
        local adapters
        mapfile -t adapters < <(_peared_adapter_candidates)
        if [ ${#adapters[@]} -eq 0 ]; then
                return
        fi
        COMPREPLY=( $(compgen -W "${adapters[*]}" -- "$cur_word") )
}

_peared_complete_log_levels() {
        local cur_word="$1"
        COMPREPLY=( $(compgen -W "debug info warn warning error" -- "$cur_word") )
}

_peared_complete_duration() {
        local cur_word="$1"
        COMPREPLY=( $(compgen -W "5s 10s 15s 30s 45s 60s 1m" -- "$cur_word") )
}

_peared()
{
        local cur prev words cword
        COMPREPLY=()
        words=("${COMP_WORDS[@]}")
        cword=$COMP_CWORD
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"

        if [ $cword -le 1 ]; then
                COMPREPLY=( $(compgen -W "adapters devices shell help" -- "$cur") )
                return
        fi

        case "${words[1]}" in
        shell)
                case "$prev" in
                --log-level)
                        _peared_complete_log_levels "$cur"
                        return
                        ;;
                --prompt)
                        return
                        ;;
                esac

                if [[ "$cur" == -* ]]; then
                        COMPREPLY=( $(compgen -W "--log-level --prompt --help -h" -- "$cur") )
                fi
                ;;
        adapters)
                if [ $cword -eq 2 ]; then
                        COMPREPLY=( $(compgen -W "list help" -- "$cur") )
                        return
                fi

                case "${words[2]}" in
                list)
                        case "$prev" in
                        --sysfs)
                                _peared_complete_dirs "$cur"
                                return
                                ;;
                        esac

                        if [[ "$cur" == -* ]]; then
                                COMPREPLY=( $(compgen -W "--sysfs --help -h" -- "$cur") )
                        fi
                        ;;
                help)
                        if [[ "$cur" == -* ]]; then
                                COMPREPLY=( $(compgen -W "--help -h" -- "$cur") )
                        fi
                        ;;
                esac
                ;;
        devices)
                if [ $cword -eq 2 ]; then
                        COMPREPLY=( $(compgen -W "scan pair connect disconnect help" -- "$cur") )
                        return
                fi

                case "${words[2]}" in
                scan)
                        case "$prev" in
                        --duration)
                                _peared_complete_duration "$cur"
                                return
                                ;;
                        --config)
                                _peared_complete_files "$cur"
                                return
                                ;;
                        --adapter)
                                _peared_complete_adapters "$cur"
                                return
                                ;;
                        esac

                        if [[ "$cur" == -* ]]; then
                                COMPREPLY=( $(compgen -W "--duration --no-sudo --adapter --config --help -h" -- "$cur") )
                        fi
                        ;;
                pair|connect|disconnect)
                        case "$prev" in
                        --config)
                                _peared_complete_files "$cur"
                                return
                                ;;
                        --adapter)
                                _peared_complete_adapters "$cur"
                                return
                                ;;
                        esac

                        if [[ "$cur" == -* ]]; then
                                COMPREPLY=( $(compgen -W "--no-sudo --adapter --config --help -h" -- "$cur") )
                        fi
                        ;;
                help)
                        if [[ "$cur" == -* ]]; then
                                COMPREPLY=( $(compgen -W "--help -h" -- "$cur") )
                        fi
                        ;;
                esac
                ;;
        help)
                if [ $cword -eq 2 ]; then
                        COMPREPLY=( $(compgen -W "adapters devices shell" -- "$cur") )
                        return
                fi
                ;;
        esac
}

complete -F _peared peared
