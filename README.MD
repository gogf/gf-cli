# GF-CLI
English | [简体中文](README_ZH.MD)

`gf-cli` is a powerful CLI tool for building [GoFrame](https://goframe.org) application with convenience.

## 1. Install

### 1). Binary
It's recommended using pre-compiled binaries:
- `Linux   (amd64)`: https://goframe.org/cli/linux_amd64/gf
- `Mac     (amd64)`: https://goframe.org/cli/darwin_amd64/gf
- `Windows (amd64)`: https://goframe.org/cli/windows_amd64/gf.exe
- [More...](https://goframe.org/cli)

After downloads, please use `gf install` command to install gf binary to system binary path.

1. `Mac`
    ```shell
    wget https://goframe.org/cli/darwin_amd64/gf && chmod +x gf && ./gf install
    ```
   > If you're using `zsh`, you might need rename your alias by command `alias gf=gf` to resolve the conflicts between `gf` and `git fetch`.
                                                                                                                          
1. `Linux` 
    ```shell
    wget https://goframe.org/cli/linux_amd64/gf && chmod +x gf && ./gf install
    ```
                                                                                                                  
1. `Windows`

    Manually download, execute it and then follow the instruction.

### 2). Manually Install

> You might need setting the goproxy to make through building.

1. Latest version
    ```
    git clone https://github.com/gogf/gf-cli 
    cd gf-cli && go build main.go && ./main install
    ```
1. Specified version
    ```
    git clone --branch v0.7.3 https://github.com/gogf/gf-cli 
    cd gf-cli && go build main.go && ./main install
    ```
1. Database `sqlite` and `oracle` are not support in `gf gen` command in default as it needs `cgo` and `gcc`, you can manually make some changes to the source codes and do the building.

## 2. Commands
```html
$ gf
USAGE
    gf COMMAND [ARGUMENT] [OPTION]

COMMAND
    env        show current Golang environment variables
    get        install or update GF to system in default...
    gen        automatically generate go files for ORM models...
    mod        extra features for go modules...
    run        running go codes with hot-compiled-like feature...
    init       initialize an empty GF project at current working directory...
    help       show more information about a specified command
    pack       packing any file/directory to a resource file, or a go file...
    build      cross-building go project for lots of platforms...
    docker     create a docker image for current GF project...
    swagger    swagger feature for current project...
    update     update current gf binary to latest one (might need root/admin permission)
    install    install gf binary to system (might need root/admin permission)
    version    show current binary version info

OPTION
    -y         all yes for all command without prompt ask
    -?,-h      show this help or detail for specified command
    -v,-i      show version information

ADDITIONAL
    Use 'gf help COMMAND' or 'gf COMMAND -h' for detail about a command, which has '...'
    in the tail of their comments.
```

## 3. FAQ

### 1). Command `gf run` returns `pipe: too many open files`

Please use `ulimit -n 65535` to enlarge your system configuration for max open files for current terminal shell session, and then `gf run`.







