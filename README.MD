# GF-CLI

## Install

It's suggested using pre-compiled binaries:
https://goframe.org/cli/

Popular binaries:
- `Linux   (amd64)`: https://goframe.org/cli/latest/linux_amd64/gf
- `Mac     (amd64)`: https://goframe.org/cli/latest/darwin_amd64/gf
- `Windows (amd64)`: https://goframe.org/cli/latest/windows_amd64/gf.exe

After downloads, please use `gf install` installing gf binary to system bin path.

## Commands
```html
$ gf
Usage   : gf [COMMAND] [OPTION]
Commands:
    ?,-?,-h,help  : this help.
    -v,-i,version : show version info.
    get           : install or update GF to system.
    init          : initialize an empty GF project at current working directory.
        [NAME]    : name for current GF project, not necessary, default name is 'gf-app'.
    install       : install gf binary to system (you may need root/admin permission).
    compile       : cross-compile go project.
        FILE          : compiling file path.
        -n, --name    : binary name.
        -v, --version : binary version.
        -a, --arch    : architecture for compiling, multiple.
            --os      : target operation system for compiling
```
