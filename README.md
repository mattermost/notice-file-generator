# notice-file-generator
Notice file generator Mattermost tool to automatically generate NOTICE file for Go, Node and Python projects.

## Get Involved

- [Join the discussion on ~Developers: DevOps](https://community.mattermost.com/core/channels/build)

## Developing

### Environment Setup

Essentials:

1. Install [Go](https://golang.org/doc/install)
2. Configure [Github Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)

Optionals:

1. Configure [config.yaml](.config/notice-file/config.yaml). Documentation is [here](.config/notice-file/config.yaml).

### Running
Execute go run with command line parameters:

```
go run ./... -c <CONFIG_FILE> -n <PROECT_NAME> -p <PATH_TO_PROJECT> -t <GITHUB_TOKEN>
```

Arguments:

| Name | Argument | Description |
| :--  | :--      | :---------- |
| Configuration File | -c <path_to_config_file> | Full path of the configuration file. |
| Name | -n <project_name> | Name of the project, tool will create temporary folders by using project name.
| Project Path | -p <project_path> | Full path of the project's root directory |
| Github Token (optional) | -t <github_pat_token> | Dependency licences will be fetched from Github, token needed to remove API rate limits. |

### Testing

Running all tests:

```shell
make test
```

## License

See [LICENSE.txt](LICENSE.txt) for license rights and limitations.


## Release

Create a tag for desired version, pipelines will create and publish the release.
