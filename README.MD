# GF-CLI

`gf-cli` is a powerful CLI tool for building [GF](https://goframe.org) application more convenience.

## 1. Install

It's recommended using pre-compiled binaries:
- `Linux   (amd64)`: https://goframe.org/cli/linux_amd64/gf
- `Mac     (amd64)`: https://goframe.org/cli/darwin_amd64/gf
- `Windows (amd64)`: https://goframe.org/cli/windows_amd64/gf.exe
- [More...](https://goframe.org/cli)

After downloads, please use `gf install` command to install gf binary to system binary path.

1. `Linux`
    ```shell
    wget https://goframe.org/cli/linux_amd64/gf && chmod +x gf && sudo ./gf install
    ```
1. `Mac`
    ```shell
    wget https://goframe.org/cli/darwin_amd64/gf && chmod +x gf && ./gf install
    ```
   > If you're using `zsh`, you might need rename your alias by command `alias gf=gf` to resolve the conflicts between `gf` and `git fetch`.
1. `Windows`

    Manually download and move it to `C:\Windows` folder, and then you can enjoy it in `cmd`.

## 2. Commands
```html
$ gf
USAGE
    gf COMMAND [ARGUMENT] [OPTION]

COMMAND
    get        install or update GF to system in default...
    gen        automatically generate go files for ORM models...
    run        running go codes with hot-compiled-like feature...
    init       initialize an empty GF project at current working directory...
    help       show more information about a specified command
    pack       packing any file/directory to a resource file, or a go file
    build      cross-building go project for lots of platforms...
    update     update current gf binary to latest one (you may need root/admin permission)
    install    install gf binary to system (you may need root/admin permission)
    version    show version info

OPTION
    -?,-h      show this help or detail for specified command
    -v,-i      show version information

ADDITIONAL
    Use 'gf help COMMAND' or 'gf COMMAND -h' for detail about a command, which has '...' in the tail of their comments.

```
