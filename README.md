# gocc [![Go Report Card](https://goreportcard.com/badge/github.com/skifli/gocc)](https://goreportcard.com/report/github.com/skifli/gocc)

- [gocc ](#gocc-)
  - [Installation](#installation)
    - [Using pre-built binaries](#using-pre-built-binaries)
    - [Running from source](#running-from-source)
  - [Usage](#usage)
  - [Configuration](#configuration)
  - [How to use the configuration file](#how-to-use-the-configuration-file)

gocc (pronounced as _**/ɡɒtʃ/**_) makes cross-compilation for [Go](https://go.dev) easy, by automating the process on all OSes. It allows for easy cofiguration of build architectures, and has pre-built binaries for all releases.

## Installation

### Using pre-built binaries

Download the binary for your OS from the [latest release](https://github.com/skifli/gocc/releases/latest). There are quick links at the top of every release for popular OSes. If you are on **Linux or macOS**, you may have to execute the following command in a shell to be able to run the binary.

```bash
chmod +x path_to_binary
```

### Running from source

Use this method if none of the pre-built binaries work on your system. It is possible that your system's architecture is different to the one that the binaries were compiled for **(AMD)**. You can check your system's architecture by viewing the value of the **`GOHOSTARCH`** environment variable.

* Make sure you have [Go](https://go.dev) installed and is in your system environment variables as **`go`**. If you do not have go installed, you can install it from [here](https://go.dev/dl/).
* Download and extract the repository from [here](https://github.com/skifli/gocc/archive/refs/heads/master.zip). Alternatively, you can clone the repository with [Git](https://git-scm.com/) by running `git clone https://github.com/skifli/gocc` in a terminal.
* Navigate into the `/src` directory of your clone of this repository.
* Run the command `go build main.go`.
* The compiled binary is in the same folder, named `main.exe` if you are on Windows, else `main`.

## Usage

```
main - gocc: Go Cross-Compiling made easy. Get more information at https://github.com/skifli/gocc.

Usage:
    main [target]

Positional Variables: 
    target   The path to the file to cross-compile. (Required)

Flags: 
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
    -d --dump      The path to the folder to dump the cross-compiled binaries in. Defaults to `build` in the cwd. The folder will be created if it does not exist.
    -c --config    The path to the config file.
```

## Configuration

An example configuration file can be found in [`/src/example.cfg`](https://github.com/skifli/gocc/blob/main/src/example.cfg).

## How to use the configuration file

All OS / Architecture combinations that are found in the config file are not compiled for. The combinations should follow the format of the output of `go tool dist list`. For example, `windows/amd64` is valid, but `windows\amd64` is not. The `*` character can also be used to specify all targets. For example, `windows/*` applies to all architectures under the OS _**Windows**_, regardless of architecture; `*/386` applies to all OSes with the architecture _**`386` (32-Bit)**_, regardless of OS.

_**NB:**_
Lines _**starting**_ with `#` are ignored when parsing the configuration file. If you provide a combination that does not exist, the program will ignore it, and _**no error will be raised**_. If the program encounters an error when parsing the file, the script will output the line the error occured on, and exit.
