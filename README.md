# gocc 

[![Go Report Card](https://goreportcard.com/badge/github.com/skifli/gocc)](https://goreportcard.com/report/github.com/skifli/gocc)
![Lines of Code](https://img.shields.io/github/languages/code-size/skifli/gocc)
[![Release Downloads](https://img.shields.io/github/downloads/skifli/gocc/total.svg)](https://github.com/skifli/gocc/releases)

- [gocc](#gocc)
  - [Installation](#installation)
    - [Using pre-built binaries](#using-pre-built-binaries)
    - [Running from source](#running-from-source)
  - [Usage](#usage)
  - [Configuration](#configuration)
    - [How the configuration file works](#how-the-configuration-file-works)
  - [Known issues](#known-issues)
    - [#1 - **loadinternal: cannot find runtime/cgo**](#1---loadinternal-cannot-find-runtimecgo)
  - [Stargazers over time](#stargazers-over-time)

gocc (pronounced as _**/ɡɒtʃ/**_) makes cross-compilation for [Go](https://go.dev) easy, by automating the process on all OSes. It allows for easy cofiguration of build architectures, and has pre-built binaries for all releases.

## Installation

### Using pre-built binaries

Pre-built binaries are made available for every `x.x` release. If you want more frequent updates, then [run from source](#running-from-source). Download the binary for your OS from the [latest release](https://github.com/skifli/gocc/releases/latest). There are quick links at the top of every release for popular OSes.

> **Note** If you are on **Linux or macOS**, you may have to execute **`chmod +x path_to_binary`** in a shell to be able to run the binary.

### Running from source

Use this method if none of the pre-built binaries work on your system, or if you want more frequent updates. It is possible that your system's architecture is different to the one that the binaries were compiled for **(AMD)**.

> **Note** You can check your system's architecture by viewing the value of the **`GOHOSTARCH`** environment variable.

* Make sure you have [Go](https://go.dev) installed and is in your system environment variables as **`go`**. If you do not have go installed, you can install it from [here](https://go.dev/dl/).
* Download and extract the repository from [here](https://github.com/skifli/gocc/archive/refs/heads/master.zip). Alternatively, you can clone the repository with [Git](https://git-scm.com/) by running `git clone https://github.com/skifli/gocc` in a terminal.
* Navigate into the `/src` directory of your clone of this repository.
* Run the command `go build main.go`.
* The compiled binary is in the same folder, named `main.exe` if you are on Windows, else `main`.

## Usage

```
usage: gocc [-h|--help] [_positionalArg_gocc_1 "<value>"] [-d|--dump "<value>"]
            [-c|--config "<value>"]

            Go Cross-Compiling made easy (v1.2.0). Get more information at
            https://github.com/skifli/gocc

Arguments:

  -h  --help                   Print help information
      --_positionalArg_gocc_1  The path to the file to cross-compile.
  -d  --dump                   The path to the folder to dump the
                               cross-compiled binaries in. Defaults to `build`
                               in the cwd. The specified folder will be created
                               if it does not exist.
  -c  --config                 The path to the config file.
```

## Configuration

Example configuration files can be found in [`/config_examples`](https://github.com/skifli/gocc/blob/main/config_examples).

### How the configuration file works

Open [`/config_examples/allow.json`](https://github.com/skifli/gocc/blob/main/config_examples/allow.json). You will see two keys - `mode`, and **`targets`**. Mode can be a string of two values - **allow**, or **disallow**. The file you have just opened has the value of `mode` set to **allow**. This means that only the OS / Architecture combinitations in the list **`targets`** will be compiled for. **`targets`** is a list of strings containing the OS / Architecture combinations, in the form outputted from the command **`go tool dist list`**. This means **`windows/amd64`** is valid, but not **`windows\amd64`**. If you ran the program with the configuration file you have open, it would **only compile** for all architectures under the OS **Windows**, and all OSes with the architecture **386 (32-Bit)**.

> **Note** The `*` character can also be used to specify all targets. For example, `windows/*` applies to all architectures under the OS _**Windows**_, regardless of architecture; `*/386` applies to all OSes with the architecture _**`386` (32-Bit)**_, regardless of OS.

Now open [`/config_examples/disallow.json`](https://github.com/skifli/gocc/blob/main/config_examples/disallow.json). You will see the file is the exact same as the previously opened file, except for the fact that `mode` is set to `disallow`. This means that all OS / Architecture combinations in the list **`targets`** will **not** be compiled for, and all combinations not in the list **will** be compiled for. If you ran the program with the configuration file you have open, it **would compile** for all OSes except for **Windows**, and for architectures except for **386 (32-Bit)**.

> **Note** If you provide a combination that does not exist, the program will ignore it, and _**no error will be raised**_. If the program encounters an error when parsing the config file, the script will output the culprit line / combination, and exit.

## Known issues

### [#1](https://github.com/skifli/gocc/issues/1) - **loadinternal: cannot find runtime/cgo**

This issue is caused by the **`runtime/cgo`** module being disabled. gocc will automatically enable **`runtime/cgo`**, but only if [**`gcc`**](https://gcc.gnu.org/) is installed. This is because **`runtime/cgo`** requires gcc to be installed in PATH, so it can be used to compile C code. Instructions for installing gcc can be found below:

* [Windows](https://github.com/danielpinto8zz6/c-cpp-compile-run/blob/HEAD/docs/COMPILER_SETUP.md#windows)
* [Linux](https://github.com/danielpinto8zz6/c-cpp-compile-run/blob/HEAD/docs/COMPILER_SETUP.md#linux)
* [macOS](https://github.com/danielpinto8zz6/c-cpp-compile-run/blob/HEAD/docs/COMPILER_SETUP.md#macos)

> **Note** On **macOS**, **`clang`** and **`gcc`** are the same, which is why the installation insutrctions for macOS say to install **`clang`**.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/skifli/gocc.svg)](https://starchart.cc/skifli/gocc)
