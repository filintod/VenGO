# This file can't be executed directly, it has to be
# loaded with 'vengo_activate <ebvironment_name>' from fish

# This script is inspired by virtualenv for Python written by
# Jannis Leidel, Carl Meyer and Brian Rosner

function deactivate --description "Deactivate a VenGO active environment"
    if not set -q VENGO_ENV
        return 0
    end
    # reset environment variables
    set -x PATH $PATH[3..(count $PATH)]
    if test -n "$_VENGO_PREV_PATH"
        set -g PATH "$_VENGO_PREV_PATH"
        set -e _VENGO_PREV_PATH
    end
    if test -n "$_VENGO_PREV_GOROOT"
        set -g GOROOT "$_VENGO_PREV_GOROOT"
        set -e _VENGO_PREV_GOROOT
    end
    if test -n "$_VENGO_PREV_GOTOOLDIR"
        set -g GOTOOLDIR "$_VENGO_PREV_GOTOOLDIR"
        set -e _VENGO_PREV_GOTOOLDIR
    end
    if test -n "$_VENGO_PREV_GOPATH"
        set -g GOPATH "$_VENGO_PREV_GOPATH"
        set -e _VENGO_PREV_GOPATH
    end

    # set an empty local fish_function_path, so fish_prompt doesn't automatically reload
    set -l fish_function_path
    functions -e fish_prompt
    functions -c _vengo_fish_prompt fish_prompt
    functions -e _vengo_fish_prompt

    set -e VENGO_ENV
    set -e VENGO_PROMPT
    if test "$argv[1]" != "just_reset"
        functions -e deactivate
    end
end

# reset environment
# this is useful if someone activate an environment while other
# environment is still active for
deactivate just_reset

# set paths
set -g VENGO_ENV "{{ .VenGO_PATH }}"
set -x VENGO_ENV $VENGO_ENV
# unset and backup old configuration
set -g _VENGO_PREV_GOROOT (go env GOROOT)
set -e GOROOT

set -g _VENGO_PREV_GOTOOLDIR (go env GOTOOLDIR)
set -e GOTOOLDIR

set -g _VENGO_PREV_GOPATH (go env GOPATH)
set -e GOPATH

# set new environment variables
set -g GOROOT "{{ .Goroot }}"
set -g GOTOOLDIR "{{ .Gotooldir }}"
set -g GOPATH "{{ .Gopath }}"
set -x GOPATH $GOPATH
set -g VENGO_PROMPT "{{ .PS1 }}"

# set the PATH
set -g PATH "$GOROOT/bin" "$GOPATH/bin" $PATH

# copy fish_prompt and overwrite it with our own
functions -c fish_prompt _vengo_fish_prompt
function fish_prompt
    # override prompt is prompt exists
    if test -n "$VENGO_PROMPT"
        printf "%s%s" "$VENGO_PROMPT" (set_color normal)
        _vengo_fish_prompt
        return
    else
        # prepend the VenGO environment name
        printf "(%s) %s" (basename "$VENGO_ENV") (set_color normal)
        _vengo_fish_prompt
    end
end
