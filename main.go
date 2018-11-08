package main

import (
	"context"
	"fmt"
	"os"
	"strings"
    "io/ioutil"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Struct githubUser representing a github user with a list of SSH public keys
// name: github user
// keys: list of ssh keys configured inside Github
type githubUser struct {
	name string
	keys []string
}

// Check if the user has any SSH public key
func (gu githubUser) hasKeys() bool {
    return len(gu.keys) >0
}


// Returns string with all the user public keys to be written on 
// authorized_keys file, one line per key
func (gu githubUser) formatOut() string {
    formatted := ""
    for numKey, key := range gu.keys {
       line := fmt.Sprintf("%s %s-%d\n", key, gu.name, numKey)
       formatted = formatted + line
    }
    return formatted
}

// Main function takes a github access token, organization, and list of teams
// authorized_keys file. Get all the users on the selected teams and all his
// keys and it writes it down to the authorized_keys file
func main() {

	// ENV vars to be send to the program:
	// GITHUB_ACCESS_TOKEN  --> User access token with permissions to the organization
	// GITHUB_ORGANIZATION  --> Github organization to take the users from
	// GITHUB_TEAMS         --> Comma-separated list of github teams to 
    //                          take the users from, if empty will take 
    //                          all the teams
	// AUTHORIZED_KEYS_FILE --> authorized_keys file where public keys 
    //                          will be written 
	githubAccessToken  := getEnvOrFail("GITHUB_ACCESS_TOKEN")
	githubOrganization := getEnvOrFail("GITHUB_ORGANIZATION")
	githubTeams        := getEnvOrDefault("GITHUB_TEAMS", "")
    home               := os.Getenv("HOME")
	authorizedKeysFile := getEnvOrDefault("AUTHORIZED_KEYS_FILE", home + "/.ssh/authorized_keys")

	sshUsers           := getUsers(githubAccessToken, githubOrganization, githubTeams)
    writeUsersToFile(sshUsers, authorizedKeysFile)
}

// Given a list of githubUsers, write the content to the 
// authorized_keys file in the correct format
func writeUsersToFile(sshUsers []githubUser, authorizedKeysFile string) {
    allUsersStr := ""
    for _, user := range sshUsers {
        if user.hasKeys() {
            allUsersStr = allUsersStr + user.formatOut()
        }
    }
    ioutil.WriteFile(authorizedKeysFile, []byte(allUsersStr) ,0)
}

// Given a token to authenticate, organization and a list of teams, 
// it returns the list of githubUsers that are inside the teams, deleting
// duplicates
func getUsers(githubAccessToken string, githubOrganization string, githubTeams string) []githubUser {

	var users = make([]githubUser, 0)

	ctx := context.Background()
	githubClient := getGithubClient(ctx, githubAccessToken)
	teamsIds := getTeamsIds(ctx, githubClient, githubOrganization, githubTeams)
	for _, teamId := range teamsIds {
		users = append(users, getTeamUsers(ctx, githubClient, teamId, users)...)
	}
	return users
}

// Given a context, github client and team id, and a list of actual users,
// returns the user of a team, deleting duplicates from actualUsers
func getTeamUsers(context context.Context, githubClient github.Client, teamId int64, actualUsers []githubUser) []githubUser {

	var users = make([]githubUser, 0)

	members, _, err := githubClient.Teams.ListTeamMembers(context, teamId, nil)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	for _, member := range members {
		if !usersContain(actualUsers, *member.Login) {
			var user githubUser
			user.keys = make([]string, 0)
			user.name = *member.Login
			keys, _, err := githubClient.Users.ListKeys(context, *member.Login, nil)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			for _, key := range keys {
				user.keys = append(user.keys, *key.Key)
			}
			users = append(users, user)
		}
	}
	return users
}

// Given a list of users and a username, checks if username exists
// in the list og users
func usersContain(actualUsers []githubUser, username string) bool {
	for _, user := range actualUsers {
		if user.name == username {
			return true
		}
	}
	return false
}

// Given a context, githubClient and a github organization, and a list of
// teams, separeted by commas, returns the ids of the given teams
func getTeamsIds(context context.Context, githubClient github.Client, githubOrganization string, githubTeams string) []int64 {
	githubTeamsArr := strings.Split(githubTeams, ",")
	teams, _, err := githubClient.Teams.ListTeams(context, githubOrganization, nil)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	var teamsIds = make([]int64, 0)
	for _, team := range teams {
		if len(githubTeams) == 0 || contains(githubTeamsArr, *team.Name) {
			teamsIds = append(teamsIds, *team.ID)
		}
	}
	return teamsIds
}

// Given a context and a github access token, returns the client to use
// the github API
func getGithubClient(context context.Context, githubAccessToken string) github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(context, ts)

	client := github.NewClient(tc)
	return *client
}

// Get the value of an environment variable or exits
func getEnvOrFail(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		fmt.Fprintf(os.Stderr, "Error: environment variable %s doesn't exist or it's empty, set it and try it again", envVar)
		os.Exit(1)
	}
	return value
}

// Get the value of an environment variable or returns a default value
func getEnvOrDefault(envVar string, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}

// Check if an array of strings contain a string
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
