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
go run ./... -c <CONFIG_FILE> -p <PATH_TO_PROJECT> -t <GITHUB_TOKEN>
```

Arguments:

| Name | Argument | Description |
| :--  | :--      | :---------- |
| Configuration File | -c <path_to_config_file> | Full path of the configuration file. |
| Project Path (optional) | -p <project_path> | Full path of the project's root directory. Current path will be used if not provided |
| Github Token (optional) | -t <github_pat_token> | Dependency licences will be fetched from Github, token needed to remove API rate limits. |

### Testing

Running all tests:

```shell
make test
```

## License

See [LICENSE.txt](LICENSE.txt) for license rights and limitations.


## Release

To trigger a release of the Notice File Generator, follow these steps:

1. **For Patch Release:** Run the following command:
    ```
    make patch
    ```
   This will release a patch change.

2. **For Minor Release:** Run the following command:
    ```
    make minor
    ```
   This will release a minor change.

3. **For Major Release:** Run the following command:
    ```
    make major
    ```
   This will release a major change.

4. **For Patch Release Candidate (RC):** Run the following command:
    ```
    make patch-rc
    ```
   This will release a patch release candidate.

5. **For Minor Release Candidate (RC):** Run the following command:
    ```
    make minor-rc
    ```
   This will release a minor release candidate.

6. **For Major Release Candidate (RC):** Run the following command:
    ```
    make major-rc
    ```
   This will release a major release candidate.
