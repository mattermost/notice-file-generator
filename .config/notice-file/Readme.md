# Notice.txt File Configuration

We are automatically generating Notice.txt by using first-level dependencies of the project. The related pipeline uses `config.yaml` stored in this folder.


## Configuration

Sample:

```
title: "Mattermost Notice File Generator"
copyright: "©2022 Mattermost, Inc.  All Rights Reserved.  See LICENSE.txt for license information."
description: "This document includes a list of open source components used in Mattermost Motice File Generator, including those that have been modified."
search:
  - "go.mod"
additionalDependencies:
  - wix
```

| Field                  | Type    | Purpose                                                                                                                  |
| :--------------------- | :------ | :----------------------------------------------------------------------------------------------------------------------- |
| title                  | string  | Field content will be used as a title of the application. See first line of `NOTICE.txt` file.                           |
| copyright              | string  | Field content will be used as a copyright message. See second line of `NOTICE.txt` file.                                 |
| description            | string  | Field content will be used as notice file description. See third line of `NOTICE.txt` file.                              |
| includeDevDependencies | boolean | If true we include devDependency section of all package.json files declared.                                             |
| additionalDependencies | array   | Optional additional dependencies. Their stanzas in the `NOTICE.txt` file should be added manually.                       |
| search                 | array   | Pipeline will search for package.json files mentioned here. Globstar format is supported ie. `packages/**/package.json`. |
