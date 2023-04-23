## Introduction
GoJump is a bastion server inspired by [KoKo](https://github.com/jumpserver/koko) but integrated backend and admin manager.

Also most code in this project is copied & modified from [KoKo](https://github.com/jumpserver/koko).

## Why GoJump
Since KoKo is just component of [JumpServer](https://github.com/jumpserver/jumpserver), it cannot run independently.

GoJump is a light JumpServer integrated ssh service and admin manager. You can use to start a bastion server fleetly without complicated configuration.

It's very appropriate for light usage.

## Features
- Support SSH protocal
- Support VS Code(dangerous)
- Once time password
- Login confirm
- Record replay based on [asciicast v2](https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md)

## Building from source
```bash
git clone https://github.com/handewo/gojump.git
cd gojump
chmod u+x build.sh
# By default, the script will delete and initial gojumpdb
./build.sh
```
## RoadMap
- Support more protocal like MySQL, PostgreSQL, Redis, etc.
- Provide RESTful api for admin manager
- Provide pretty ui of admin manager
- Support MFA authentication

## Tech Stack
- Database: [genji](https://github.com/genjidb/genji), is unstable now but very convenient for developing.