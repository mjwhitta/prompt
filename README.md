# Prompt

[![Yum](https://img.shields.io/badge/-Buy%20me%20a%20cookie-blue?labelColor=grey&logo=cookiecutter&style=for-the-badge)](https://www.buymeacoffee.com/mjwhitta)

[![Go Report Card](https://goreportcard.com/badge/github.com/mjwhitta/prompt?style=for-the-badge)](https://goreportcard.com/report/github.com/mjwhitta/prompt)
![License](https://img.shields.io/github/license/mjwhitta/prompt?style=for-the-badge)

## What is this?

A really dumb prompt. This works enough for me to simulate shell
environments for other projects. It has no tab completion. It simply
allows for processing user input with the following features;

- Adjusts to current terminal width
- Command history
    - Customizable history size
- Custom prompt
    - Color support
    - Multi-line support
- Probably some hidden ~~bugs~~features

## How to install

Open a terminal and run the following:

```
$ go get -u github.com/mjwhitta/prompt
```

## Usage

See [example](./example/main.go) for a simple bash-like prompt.

## Links

- [Source](https://github.com/mjwhitta/prompt)

## TODO

- Add support for custom key-bindings?
- Add support for tab-completion maybe?
