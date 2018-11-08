# github-ssh-keys
Creates an authorized_keys file with the public SSH keys of the users on the selected Github teams inside one organization. This allows to manage SSH authorized_keys file from Github, associating instances with teams allowed to login.

## Parametrization
Next env variables must be set:
* GITHUB_ACCESS_TOKEN - Github access token that have permissions to read the organization info
* GITHUB_ORGANIZATION - Github organization that contains teams with the users

Next env variables can be set:
* GITHUB_TEAMS - Comma separated list of github teams, the users keys of this will be writed to the authorized_keys file. If it's not specified all the teams of the organization will be taken.
* AUTHORIZED_KEYS_FILE - authorized_keys file to write the SSH public keys.

## Running it
Execution example with docker, best option is to add this command to cron:
```
docker run -v /home/user/.ssh/authorized_keys:/home/user/.ssh/authorized_keys -e GITHUB_ACCESS_TOKEN=***** -e GITHUB_ORGANIZATION=your_org -e GITHUB_TEAMS=frontend,backend,operations -e AUTHORIZED_KEYS_FILE=/home/user/.ssh/authorized_keys --user=$(id -u) orimarti/github-ssh-keys
```
