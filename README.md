[![Go Report Card](https://goreportcard.com/badge/github.com/unly/wow-addon-updater)](https://goreportcard.com/report/github.com/unly/wow-addon-updater)
[![License](https://img.shields.io/badge/license-MIT-green)](https://github.com/unly/wow-addon-updater/blob/master/LICENSE)
[![CI Status](https://github.com/unly/wow-addon-updater/workflows/CI/badge.svg)](https://github.com/unly/wow-addon-updater/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/unly/wow-addon-updater/branch/master/graph/badge.svg?token=HZ0DG1CL6E)](https://codecov.io/gh/unly/wow-addon-updater)

# WoW-Addon-Updater

Currently supported AddOn sources:
* [github.com](https://github.com/)
* [tukui.org](https://www.tukui.org/)

## Run the Updater

For windows simply run the `updater.exe` file and the terminal with the output should pop up.
For linux and macOS simply run `./updater` from the terminal.
By default the updater will look for a configuration file `config.yaml` in the same directory as the application.
To use a different configuration file path run the application with the `-c` flag, e.g. `-c path/to/config` to overwrite the default.

## Configuration

A configuration file contains the path to the interface directory on your system as well as the list of addons.
This works for classic and retail identical.

```yaml
classic:
    path: path/to/classic/interface/directory
    addons:
    - https://www.tukui.org/classic-addons.php?id=1
    - https://github.com/AeroScripts/QuestieDev
retail:
    path: path/to/retail/interface/directory
    addons: []
```