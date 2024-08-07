package common

import "strings"

func LinuxPathJoin(paths ...string) string {
	joinedPath := ""
	for _, pathPartName := range paths {
		if joinedPath == "" { /* just add for the first item */
			joinedPath = strings.ReplaceAll(pathPartName, "./", "")
			continue
		}

		cleanedPath := strings.ReplaceAll(pathPartName, "./", "/")

		if joinedPath[len(joinedPath)-1] != '/' && joinedPath[len(joinedPath)-1] != '\\' {
			joinedPath += "/"
		}

		if len(cleanedPath) == 0 {
			continue
		}
		if cleanedPath[0] == '/' || cleanedPath[0] == '\\' {
			joinedPath += cleanedPath[1:]
		} else {
			joinedPath += cleanedPath
		}
	}

	return joinedPath
}
